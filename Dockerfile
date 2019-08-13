FROM golang:1.12-alpine as buildStage
LABEL maintainer 'Daniel M. Lambea <dmlambea@gmail.com>'

RUN apk --no-cache add binutils git

COPY .version $GOPATH/staging/
COPY go.* $GOPATH/staging/
COPY *.go $GOPATH/staging/
COPY internal $GOPATH/staging/internal

## Build and strip the final binary
RUN cd $GOPATH/staging/ \
    && VER=`cat .version` \
    && MOD=`grep '^module' go.mod | sed 's/^[[:space:]]*module[[:space:]]\+//' | sed 's/[[:space:]]*$//'` \
    && go build -ldflags "-X $MOD/internal/version.Number=$VER" -o /tmp/drone-kube . \
    && strip /tmp/drone-kube

######################################################################

FROM alpine:3.9
LABEL maintainer 'Daniel M. Lambea <dmlambea@gmail.com>'

RUN apk --no-cache add ca-certificates

COPY --from=buildStage /tmp/drone-kube /bin/

ENTRYPOINT ["/bin/drone-kube"]
