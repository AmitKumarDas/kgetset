# Build the operator binary
FROM golang:1.12.5 as builder

WORKDIR /workspace

# copy build manifests
COPY Makefile Makefile

# copy go modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# ensure vendoring is up-to-date by running make vendor in your local
# setup
#
# we cache the vendored dependencies before building and copying source
# so that we don't need to re-download when source changes don't invalidate
# our downloaded layer
RUN make vendor-cache

# copy go source code
COPY *.go ./

ENTRYPOINT ["make","test"]
