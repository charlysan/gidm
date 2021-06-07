ARG GO_VERSION=1.15

# Step 1 - Build app binary
FROM golang:${GO_VERSION}-alpine AS builder

RUN mkdir -p /app
ADD . /app
WORKDIR /app

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /gidm main.go

# Step 2 - Build image
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /gidm /gidm

USER 10001:10001

ENTRYPOINT ["/gidm"]