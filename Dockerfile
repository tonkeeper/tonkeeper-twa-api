FROM docker.io/library/golang:1.20-alpine as builder

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
RUN apk add make
RUN make gen
RUN make build

FROM docker.io/library/alpine:latest
RUN apk --no-cache add ca-certificates \
  && update-ca-certificates

COPY --from=builder /go/src/github.com/tonkeeper/tonkeeper-twa-api/bin/twa-api .

ENTRYPOINT ["/twa-api"]
