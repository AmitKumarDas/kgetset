# Build the operator binary
FROM golang:1.12.5 as builder

WORKDIR /workspace

# copy build manifests
COPY Makefile Makefile

# copy go modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# ensure vendoring is up-to-date
# cache deps before building and copying source so that we don't need to 
# re-download as much and so that source changes don't invalidate our 
# downloaded layer
RUN make vendor

# copy go source code
COPY case_1_test.go case_1_test.go
COPY specs.go specs.go

ENTRYPOINT ["make test"]
