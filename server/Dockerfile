FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache curl
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-chat ./cmd
RUN ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') && \
    curl -fsSL "https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-${ARCH}.tar.gz" \
    | tar -xz -C /usr/local/bin

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 1001 appuser
WORKDIR /app
COPY --from=builder /go-chat .
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY db ./db
USER appuser
EXPOSE 8080
CMD ["./go-chat"]
