# Drone Kubernetes
This drone kubernetes plugin does the equivalent of: 

```
kubectl apply -f deployment.yaml
```

If the deployment does not exists, it will be created.

The advantages of this plugin is that the ```deployment.yaml``` file can be a template file.  We are able to substitute values like ```{{ build.number }}``` inside the file so you can update docker image names. 

Basic example: 

```yaml
pipeline:
  deploy:
  	 image: dmlambea/drone-kube
     template: deployment.yaml
```

Example configuration with non-default namespace:

```diff
pipeline:
  kube:
  	image: dmlambea/drone-kube
    template: deployment.yaml
+   namespace: mynamespace
```

You can also specify the server in the configuration as well.  It could alternatively be specified as an environment variable as shown in the next section. 

```diff
pipeline:
  kubernetes:
  	image: dmlambea/drone-kube
    template: deployment.yaml
+   namespace: mynamespace
+   server: https://10.93.234.28:6433
```

## Secrets

The kube plugin supports reading credentials from the Drone secret store.  This is strongly recommended instead of storing credentials in the pipeline configuration in plain text.  

Authentication against the Kubernetes API server is allowed by providing a custom *kubeconfig* filename with __KUBE_CONFIG__. If this file is not provided, then:

1. The __KUBE_SERVER__ envvar containing the server url for your Kubernetes API server is mandatory.  E.g.: https://10.1.0.1
2. A base64-encoded CA cert can be provided with __KUBE_CA__, otherwise the default Kubernetes */var/run/secrets/kubernetes.io/serviceaccount/ca.crt* file is used. More on this in the ServiceAccount section below.
3. If no __KUBE_CLIENT_CERT__ or __KUBE_CLIENT_KEY__ files are provided, then the token in __KUBE_TOKEN__ is mandatory.

The base64-encoded CA in __KUBE_CA__ can be obtained by running the command:  

```
export KUBE_CA=$(cat ca.pem | base64)
``` 

## Template Reference

You can substitute the following values between ```{{ }}``` in your deployment template 

repo.owner
: repository owner

repo.name
: repository name

build.status
: build status type enumeration, either `success` or `failure`

build.event
: build event type enumeration, one of `push`, `pull_request`, `tag`, `deployment`

build.number
: build number

build.commit
: git sha for current commit

build.branch
: git branch for current commit

build.tag
: git tag for current commit

build.ref
: git ref for current commit

build.author
: git author for current commit

build.link
: link the the build results in drone

build.created
: unix timestamp for build creation

build.started
: unix timestamp for build started

# Template Function Reference

uppercasefirst
: converts the first letter of a string to uppercase

uppercase
: converts a string to uppercase

lowercase
: converts a string to lowercase. Example `{{lowercase build.author}}`

datetime
: converts a unix timestamp to a date time string. Example `{{datetime build.started}}`

success
: returns true if the build is successful

failure
: returns true if the build is failed

truncate
: returns a truncated string to n characters. Example `{{truncate build.sha 8}}`

urlencode
: returns a url encoded string

since
: returns a duration string between now and the given timestamp. Example `{{since build.started}}`
	
## Mounting a ServiceAccount for obtaining the API server CA automatically

Official Drone is not currently able to mount the default service account in pods. You can get an enhanced Drone with this feature by cloning/forking my repo at https://github.com/dmlambea/drone.

However, the master branch in my repo includes other interesting features, such as VolumeSecret volumes to be mounted in containers. This allows the __KUBE_CLIENT_KEY__ file to be easily mounted into the plugin. Example:

```yaml
kind: pipeline
name: default

steps:

  ...

  - name: kubernetes-deploy
    image: dmlambea/drone-kube
    settings:
      automountServiceAccountToken: true
    volumes:
      - name: clientCert
        path: "/etc/ssl/client"
    environment:
      KUBE_SERVER: "https://kubernetes.default.svc"
      KUBE_CLIENT_CERT: "/etc/ssl/client/cert.pem"
      KUBE_CLIENT_KEY: "/etc/ssl/client/key.pem"
      KUBE_TEMPLATE: "build/k8s/deployment.tpl"
      KUBE_NAMESPACE: "default"

  ...

volumes:
  - name: clientCert
    secret:
      name: myKubernetesSecret
      items:
        - key: client-cert
          path: cert.pem
        - key: client-key
          path: key.pem
```

The *automountServiceAccountToken* setting tells Drone to mount the default serviceaccount token, so a ca.crt is available at its default mountpoint */var/run/secrets/kubernetes.io/serviceaccount/ca.crt*. The volume *clientCert* is similar to Kubernetes' secret volumes.
