FROM alpine:latest

RUN apk update && apk add go

ENV GOPATH=/go PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /inspector

COPY ./ ./

CMD ["go", "run", "net.go", "inspector.go", "main.go"]