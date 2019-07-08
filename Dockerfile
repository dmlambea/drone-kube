FROM alpine:3.9
LABEL maintainer 'Daniel M. Lambea <dmlambea@gmail.com>'

RUN apk --no-cache add ca-certificates

COPY drone-kube /bin/

ENTRYPOINT ["/bin/drone-kube"]
