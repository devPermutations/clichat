# Multi-stage build for Linux container

FROM golang:1.24 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build static binary (CGO disabled) for Linux amd64
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags "-s -w" -o /out/clichat ./cmd/clichat

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates
WORKDIR /app
# Install binary in PATH so mounting /app doesn't hide it
COPY --from=builder /out/clichat /usr/local/bin/clichat
# Default workdir contains .env and DB (mounted at runtime)
ENTRYPOINT ["clichat"]
CMD ["chat"]


