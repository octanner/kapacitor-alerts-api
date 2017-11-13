#!/bin/sh

go get github.com/gin-gonic/gin
cd /go/src/kapacitor-alerts-api/
go build server.go

