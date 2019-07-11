FROM golang:1.12-alpine as buildStage
LABEL maintainer 'Daniel M. Lambea <dmlambea@gmail.com>'

RUN apk --no-cache add binutils git

COPY go.* $GOPATH/staging/
COPY *.go $GOPATH/staging/

## Build and strip the final binary
RUN cd $GOPATH/staging/ && \
    go build -o /tmp/drone-kube . && \
    strip /tmp/drone-kube

######################################################################

FROM alpine:3.9
LABEL maintainer 'Daniel M. Lambea <dmlambea@gmail.com>'

RUN apk --no-cache add ca-certificates

COPY --from=buildStage /tmp/drone-kube /bin/

ENTRYPOINT ["/bin/drone-kube"]
