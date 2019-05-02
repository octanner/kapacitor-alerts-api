#!/bin/sh

go get github.com/gin-gonic/gin
go get github.com/jmoiron/sqlx
go get github.com/lib/pq
go get gopkg.in/guregu/null.v3/zero

cd /go/src/kapacitor-alerts-api/
go build -o server