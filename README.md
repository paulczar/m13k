# m13k - Mutating Webhook

This is a quick project to create a [Kubernetes Mutating Admission Controller Webhook](https://kubernetes.io/blog/2019/03/21/a-guide-to-kubernetes-admission-controllers/#example-writing-and-deploying-an-admission-controller-webhook) that simply passes the resource through a CLI tool to do the mutation.

This allows you to use tools like [ytt](https://get-ytt.io/) or [kustomize](https://github.com/kubernetes-sigs/kustomize) to modify resources as they're submitted to the Kubernetes API. This allows Kubernetes to ensure that a resource has certain labels or annotations, or even add a sidecar to a pod.

## Running

### Locally

You can test this out without running it in Kubernetes

Create TLS keypair:

```bash
openssl genrsa -out scratch/server.key 2048
openssl ecparam -genkey -name secp384r1 -out scratch/server.key
openssl req -new -x509 -sha256 -key scratch/server.key \
  -out scratch/server.pem -days 3650
```

Run `m13k` and tell it to mutate using `ytt`:

```bash
go run main.go serve --cert scratch/server.pem --key scratch/server.key --command ytt -- -o json -f - -f ./examples/ytt.yaml
```

Send a secret through and see that it comes back mutated:

> Note: this is actually an admission review request containing the secret, to emulate what really happens on the cluster.

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data @./examples/admission-review.json \
  -k https://localhost:8443/mutate | jq .response
  ```

The output of which should show an allowed patch response:

```json
{
  "uid": "11235d45-5687-4659-996d-e8f27ba7593d",
  "allowed": true,
  "status": {
    "metadata": {},
    "message": "Success"
  },
  "patch": "W3sib3AiOiJhZGQiLCJwYXRoIjoiL21ldGFkYXRhL2xhYmVscyIsInZhbHVlIjp7Im0xM2siOiJ0cnVlIn19XQ==",
  "patchType": "JSONPatch"
}

```

## Deploy

```bash
kubectl create ns m13k
kubectl apply -n m13k -f deploy/manifests.yaml
```


## Deploy via Helm

Generate certificates and save them in `deploy/helm/m13k/files` as `ca.pem`, `cert.pem` and `key.pem` make sure you set the subject to make the final service name.

This will do it for you if you don't want to mess around with `openssl` commands:

```
docker run -e SSL_SUBJECT="m13k.m13k.svc" \
  -v $(pwd)/deploy/helm/m13k/files:/certs \
  paulczar/omgwtfssl
sudo chown $USER:$USER deploy/helm/m13k/files/*
```

Deploy:

```
kubectl create ns m13k
helm install m13k --namespace m13k deploy/helm/m13k
```
