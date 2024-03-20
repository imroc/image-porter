# image-porter

An image syncing tool.

## Installation

```bash
go install github.com/imroc/image-porter@latest
```

## Usage

Use `image-porter` to sync images between different registries:

```bash
image-porter config.yaml
```

a config file is required, for example:

```yaml
default:
  tagFilter:
    regex: v?\d+(\.\d+){0,2}
images:
  - from: registry.k8s.io/ingress-nginx/kube-webhook-certgen
    to: docker.io/k8smirror/ingress-nginx-kube-webhook-certgen
    tagFilter:
      regex: ^v.*$
  - from: registry.k8s.io/ingress-nginx/opentelemetry
    to: docker.io/k8smirror/ingress-nginx-opentelemetry
    tagFilter:
      regex: ^v\w+-\w+$
```
