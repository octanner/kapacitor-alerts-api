FROM golang:1.21.0-alpine

RUN apk update
RUN apk add openssl ca-certificates git
RUN mkdir -p /go/src/kapacitor-alerts-api
WORKDIR /go/src/kapacitor-alerts-api
ADD . .
ENV GO111MODULE on
RUN go build .
CMD ["./kapacitor-alerts-api"]

