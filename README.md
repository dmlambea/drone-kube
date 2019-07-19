# Drone Kube

Drone plugin to update kubernetes objects, currently supporting apps/v1.Deployment and batch/v1beta1.CronJob.

This is a forked version from vallard/drone-kube, with recent Kubernetes v1.14 libs and support for kubeconfig files and client cert authentication.

See the [DOC](DOCS.md) file for usage. 

# Build instructions

Just use the provided *Dockerfile*:

```console
docker build -t mytag/drone-kube:latest .
```
