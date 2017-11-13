FROM golang:1.9-alpine

RUN apk update
RUN apk add openssl ca-certificates git

RUN mkdir -p /go/src/kapacitor-alerts-api
ADD server.go  /go/src/kapacitor-alerts-api/server.go

ADD build.sh /build.sh
RUN chmod +x /build.sh
RUN /build.sh

ADD start.sh /start.sh
RUN chmod +x /start.sh
CMD "/start.sh"


