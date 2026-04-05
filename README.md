# TempChat

[![CI](https://github.com/jkmoona/go-chat/actions/workflows/ci.yml/badge.svg)](https://github.com/jkmoona/go-chat/actions/workflows/ci.yml)

Disposable chat rooms. Set a TTL, share the link, talk. When time's up, the room and everything in it is gone. Messages are never written to disk.

Live: [tempchatgo.up.railway.app](https://tempchatgo.up.railway.app)

## Stack

| | |
| --- | --- |
| Backend | Go 1.23, Gin, gorilla/websocket |
| Database | PostgreSQL 15 |
| Auth | JWT (access + refresh tokens), bcrypt |
| Frontend | Vue 3, TypeScript, Tailwind CSS |
| Infra | Docker, Caddy, Railway |
| CI | GitHub Actions |

Room metadata lives in Postgres. Messages live only in memory and are broadcast over WebSocket — nothing is persisted.

## Features

- TTL between 15 minutes and 24 hours, extendable by the room creator
- Optional 4-digit PIN lock (bcrypt-hashed)
- Guest access without an account
- Live presence list and countdown bar
- Room creator can kick users, extend TTL, or delete the room
- Empty rooms auto-expire after 5 minutes of inactivity
- WebSocket heartbeat to detect and clean up silent disconnects

## Dev setup

```sh
# backend
cd server
cp .env.example .env
docker compose up --build

# frontend
cd client
cp .env.example .env
npm install
npm run dev
```

## Tests

```sh
cd server
go test -race ./...
```

## Teardown

```sh
cd server
docker compose down -v
```

## License

[MIT](LICENSE)
