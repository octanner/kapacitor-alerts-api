FROM quay.octanner.io/base/oct-golang

RUN apk update
RUN apk add openssl ca-certificates git

RUN mkdir -p /go/src/kapacitor-alerts-api
ADD . /go/src/kapacitor-alerts-api/

ADD build.sh /build.sh
RUN chmod +x /build.sh
RUN /build.sh

ADD start.sh /start.sh
RUN chmod +x /start.sh
CMD "/start.sh"
