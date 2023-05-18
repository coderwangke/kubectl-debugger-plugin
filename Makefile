.PHONY: all
all: linux

CUR := $(shell pwd)

linux:
	rm -f $(CUR)/_ouput/kubectl-debugger
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(CUR)/_output/kubectl-debugger main.go
