FROM golang:1.22-alpine3.19 AS build_deps

RUN apk add --no-cache git

WORKDIR /workspace

COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_deps AS build

COPY . .

RUN CGO_ENABLED=0 go build -o image-porter -ldflags '-w -extldflags "-static"' .

FROM alpine:3.19

VOLUME /etc/image-porter

RUN apk add --no-cache ca-certificates

COPY --from=build /workspace/image-porter /usr/local/bin/image-porter

CMD ["image-porter", "/etc/image-porter/config.yaml"]
