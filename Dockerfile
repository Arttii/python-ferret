FROM golang:1.18.2-bullseye AS builder
 
# Install git.
# Git is required for fetching the dependencies.
# Make is requiered for build.
 
RUN apt-get update && apt-get install -y make bash

WORKDIR /go/src/github.com/arttii/pyfer/pferret/lib


ADD pferret/lib/go.sum .
ADD pferret/lib/go.mod .
RUN go mod download -x
RUN go list -f '{{.Path}}/...' -m all | tail -n +2

 
COPY  pferret/lib/go.mod pferret/lib/go.sum Makefile pferret/lib/*.go ./
 

 
RUN go build -buildmode c-shared -o libferret.so

RUN go build -buildmode c-shared -o libferret.dll

FROM python:3.11.2-slim-bullseye as python


WORKDIR /opt/pyfer

COPY . /opt/pyfer
COPY --from=builder /go/src/github.com/arttii/pyfer/pferret/lib/libferret.so /opt/pyfer/pferret/lib/libferret.so
COPY --from=builder /go/src/github.com/arttii/pyfer/pferret/lib/libferret.dll /opt/pyfer/pferret/lib/libferret.dll
RUN ls /opt/pyfer/pferret/lib/ 
RUN python setup.py bdist_wheel