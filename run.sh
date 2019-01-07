#!/bin/bash
go clean
go build main.go handlers.go utils.go
./main
