FROM golang:1.18.2-bullseye AS builder
 
# Install git.
# Git is required for fetching the dependencies.
# Make is requiered for build.
 
RUN apt-get update && apt-get install -y make bash

WORKDIR /go/src/github.com/arttii/pyfer/pferret/lib

ADD pferret/lib/go.mod .
ADD pferret/lib/go.sum .
RUN go mod download -x
RUN go list -f '{{.Path}}/...' -m all | tail -n +2

 
COPY  pferret/lib/go.mod pferret/lib/go.sum Makefile pferret/lib/*.go ./
 

 
RUN go build -buildmode c-shared -o libferret.so


FROM python:3.11.2-slim-bullseye as python


WORKDIR /opt/pyfer

COPY . /opt/pyfer
COPY --from=builder /go/src/github.com/arttii/pyfer/pferret/lib/libferret.so /opt/pyfer/pferret/lib/libferret.so
RUN ls /opt/pyfer/pferret/lib/ 
RUN python setup.py bdist_wheel