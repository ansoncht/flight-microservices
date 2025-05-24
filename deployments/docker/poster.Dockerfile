FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
RUN go mod verify

COPY pkg ./pkg
COPY cmd/poster ./cmd/poster
COPY internal/poster ./internal/poster
COPY configs/poster-config.yaml ./configs/poster-config.yaml

RUN go build -ldflags="-w -s" -o /app/bin/poster ./cmd/poster

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/bin/poster .
COPY --from=builder app/configs ./configs
CMD ["./poster"]
