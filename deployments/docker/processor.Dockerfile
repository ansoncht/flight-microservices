FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
RUN go mod verify

COPY pkg ./pkg
COPY cmd/processor ./cmd/processor
COPY internal/processor ./internal/processor
COPY configs/processor-config.yaml ./configs/processor-config.yaml

RUN go build -ldflags="-w -s" -o /app/bin/processor ./cmd/processor

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bin/processor .
COPY --from=builder app/configs ./configs
CMD ["./processor"]
