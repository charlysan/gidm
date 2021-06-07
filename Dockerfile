ARG GO_VERSION=1.15
ARG ALPINE_VERSION=3.13.0

FROM golang:${GO_VERSION}-alpine AS builder

RUN mkdir -p /app
ADD . /app
WORKDIR /app

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /gidm main.go

# Create alpine runtime image
FROM alpine:${ALPINE_VERSION} as app

COPY --from=builder /gidm /gidm

USER 1000

ENTRYPOINT ["/gidm"]