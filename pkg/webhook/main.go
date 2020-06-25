package webhook

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/golang/glog"
)

var mutateCommand string

var mutateArgs []string

func init() {
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
		return
	}

	// run mutateCommand
	log.Printf("sending '%s' as stdin in to\n", string(body))
	log.Printf("mutateCommand - %s %s", mutateCommand, strings.Join(mutateArgs, " "))
	runCmd := exec.Command(mutateCommand, mutateArgs...)
	stdin, err := runCmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, string(body))
	}()

	out, err := runCmd.CombinedOutput()
	if err != nil {
		log.Printf("error: %s", out)
		log.Fatal(err)
	}

	// send response back
	// log.Printf("output - %s", out)
	io.WriteString(w, string(out))
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
