package webhook

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/mattbaird/jsonpatch"
	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var mutateCommand string
var admissionReview admission.AdmissionReview
var mutateArgs []string

func init() {
}

func admissionResponse(old, new []byte) (admission.AdmissionResponse, error) {
	patch, err := jsonpatch.CreatePatch(old, new)
	if err != nil {
		return admission.AdmissionResponse{}, err
	}
	patchBytes, _ := json.Marshal(patch)
	glog.Infof("Created patch - %s", string(patchBytes))
	response := admission.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Patch:   patchBytes,
		Allowed: true,
		PatchType: func() *admission.PatchType {
			pt := admission.PatchTypeJSONPatch
			return &pt
		}(),
		Result: &metav1.Status{
			Message: "Success",
		},
	}
	resp, _ := json.Marshal(response)
	glog.Infof("Created response - %s", resp)
	return response, err
}

func processBody(body []byte) ([]byte, error) {
	// var admissionRequest admission.AdmissionRequest
	var err error
	err = json.Unmarshal(body, &admissionReview)
	if err != nil {
		err := yaml.Unmarshal(body, &admissionReview)
		if err != nil {
			return nil, err
		}
	}
	object, err := json.Marshal(admissionReview.Request.Object)
	// admissionRequest = *admissionReview.Request
	return object, err

}

// Health responds OK
func health(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "OK\n")
}

// Mutate returns what it gets.
func mutate(w http.ResponseWriter, req *http.Request) {
	// fetch request body
	var err error
	var body []byte
	if req.Body != nil {
		if data, err := ioutil.ReadAll(req.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		io.WriteString(w, "{}")
		return
	}

	object, err := processBody(body)
	if err != nil {
		glog.Errorf("Failed to process %s", err)
		http.Error(w, "failed to decode", http.StatusBadRequest)
	}

	// run mutateCommand
	log.Printf("sending '%s' as stdin in to\n", string(object))
	log.Printf("mutateCommand - %s %s", mutateCommand, strings.Join(mutateArgs, " "))
	runCmd := exec.Command(mutateCommand, mutateArgs...)
	stdin, err := runCmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, string(object))
	}()

	out, err := runCmd.CombinedOutput()
	if err != nil {
		log.Printf("error: %s", out)
		log.Fatal(err)
	}

	patch, err := admissionResponse(object, out)
	if err != nil {
		log.Printf("error: %s", string(patch.Patch))
		log.Fatal(err)
	}
	admissionReview.Response = &patch
	response, err := json.Marshal(admissionReview)
	// send response back
	// log.Printf("output - %s", out)
	io.WriteString(w, string(response))
}

// Serve webhook
func Serve(port, cert, key, cmd string, args []string) {
	mutateCommand = cmd
	mutateArgs = args
	http.HandleFunc("/health", health)
	http.HandleFunc("/mutate", mutate)
	err := http.ListenAndServeTLS(port, cert, key, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
