#!/bin/sh

go build -o godin cmd/odin/main.go
#go build -o bin/godin cmd/odin/main.go
go build -o bin/rss cmd/rss/main.go