FROM quay.octanner.io/base/oct-golang

RUN apk update
RUN apk add openssl ca-certificates git

RUN mkdir -p /go/src/kapacitor-alerts-api
ADD . /go/src/kapacitor-alerts-api/

WORKDIR /go/src/kapacitor-alerts-api
ENV GO111MODULE on
RUN go build .
CMD "./kapacitor-alerts-api"
