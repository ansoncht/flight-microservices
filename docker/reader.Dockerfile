FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
RUN go mod verify

COPY pkg ./pkg
COPY cmd/reader ./cmd/reader
COPY internal/reader ./internal/reader
COPY configs/reader-config.yaml ./configs/reader-config.yaml

RUN go build -ldflags="-w -s" -o /app/bin/reader ./cmd/reader

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bin/reader .
COPY --from=builder app/configs ./configs
CMD ["./reader"]
