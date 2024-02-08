FROM alpine:latest

RUN apk update && apk add --no-cache go delve build-base gcc libc6-compat

ENV GOPATH=/go PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /inspector


CMD ["/bin/sh", "-c", "sleep infinity"]
#CMD ["go", "run", "net.go", "inspector.go", "main.go"]
