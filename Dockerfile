# Build the operator binary
FROM golang:1.12.5 as builder

WORKDIR /workspace

# copy go modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# ensure vendoring is up-to-date by running make vendor in your local
# setup
#
# we cache the vendored dependencies before building and copying source
# so that we don't need to re-download when source changes don't invalidate
# our downloaded layer
RUN go mod download
RUN go mod vendor

# copy build manifests
COPY Makefile Makefile

# copy go source code
COPY util/ util/
COPY cmd/ cmd/
COPY hello/ hello/
COPY unstruct/ unstruct/
COPY *.go ./

# build kgetset binary
RUN make bins

# Use distroless as minimal base image to package the final binary
#
# Refer to https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/static:latest

WORKDIR /

COPY --from=builder /workspace/kgetset /

ENTRYPOINT ["/kgetset"]
