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
COPY util/ util/
COPY *.go ./

# build the binary
RUN make lbins
RUN /bin/bash -c 'ls -la; chmod +x /workspace/kgetset.linux; ls -la'

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest

WORKDIR /
COPY --from=builder /workspace/kgetset.linux /

ENTRYPOINT ["/kgetset.linux"]
