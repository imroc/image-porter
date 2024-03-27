# image-porter

An image sync tool based on declarative configuration file, which can specify the source and destination of the images, and filter the tag list that needs to be synchronized, only sync newly pushed image tags if image have synced before.

You can use crontab / cronjob to sync images periodicly, so that keep image tags update to date.

## Installation

```bash
go install github.com/imroc/image-porter@latest
```

## Usage

Use `image-porter` to sync images between different registries:

```bash
image-porter config.yaml
```

A config file is required, for example:

```yaml
default:
  tagFilter:
    regex: ^v?\d+(\.\d+){0,2}$
images:
  - from: registry.k8s.io/ingress-nginx/kube-webhook-certgen
    to: docker.io/k8smirror/ingress-nginx-kube-webhook-certgen
    tagFilter:
      regex: ^v.*$
  - from: registry.k8s.io/ingress-nginx/opentelemetry
    to: docker.io/k8smirror/ingress-nginx-opentelemetry
    tagFilter:
      regex: ^v\w+-\w+$
  - from: registry.k8s.io/defaultbackend-amd64
    to: docker.io/k8smirror/defaultbackend-amd64
    tagFilter:
      regex: ^.*$
  - from: registry.k8s.io/ingress-nginx/controller
    to: docker.io/k8smirror/ingress-nginx-controller
```

## Example: Sync Images with Kubernetes CronJob

Checkout [this directory](./examples/kubernetes-cronjob) for how to sync images with Kubernetes CronJob (Use kustomize to manage manifests).
