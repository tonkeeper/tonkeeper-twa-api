FROM golang:1.20-bullseye as builder

WORKDIR /go/src/github.com/tonkeeper/tonkeeper-twa-api/

COPY go.mod .
COPY go.sum .

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd cmd
COPY pkg pkg
COPY Makefile .

# Build
RUN make gen
RUN make build

FROM ubuntu:20.04 as runner
RUN apt-get update && \
    apt-get install -y openssl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/src/github.com/tonkeeper/tonkeeper-twa-api/bin/twa-api .

ENTRYPOINT ["/twa-api"]