
FROM golang:1.25 AS backend-builder
WORKDIR /app
RUN apt-get update && apt-get install --no-install-recommends -y git
COPY go.mod go.sum ./server/
WORKDIR /app/server
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-chat ./cmd

FROM alpine:3.20
RUN apk add --no-cache postgresql-client curl ca-certificates
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz \
    | tar -xz -C /usr/local/bin
WORKDIR /app
COPY --from=backend-builder /go-chat .
COPY db ./db
EXPOSE 8080
CMD ["./go-chat"]
# CMD migrate -path ./db/migrations -database "$DATABASE_URL" up && ./go-chat
