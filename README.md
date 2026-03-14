# TempChat

[![CI](https://github.com/jkmoona/go-chat/actions/workflows/ci.yml/badge.svg)](https://github.com/jkmoona/go-chat/actions/workflows/ci.yml)

Disposable chat rooms that self-destruct. Pick a TTL, share the link, talk. When time's up, the room and everything in it is gone. Messages are never saved to disk.

[tempchatgo.up.railway.app](https://tempchatgo.up.railway.app)

## What it does

- Rooms automatically expire after a set time (15 minutes to 24 hours), and empty rooms are removed after 5 minutes of inactivity
- Optionally, rooms can be locked with a 4-digit PIN (securely hashed with bcrypt)
- Anyone with the link can join as a guest without signing up
- Live typing indicators, who's-online list, and a countdown timer as the room approaches expiry
- Auth uses JWT access + refresh tokens in HTTP-only cookies

## How it's built

```text
  Vue 3 SPA ──── Caddy ──── Go (Gin + WebSocket) ──── PostgreSQL
                         all on Railway
```

Room metadata (name, TTL, PIN hash) is in Postgres. Messages only exist in memory — they're broadcast over WebSocket and never touch the database.

| | |
| --- | --- |
| Backend | Go 1.23, Gin, gorilla/websocket |
| Database | PostgreSQL 15 |
| Auth | JWT (access + refresh), bcrypt |
| Frontend | Vue 3, TypeScript, Tailwind CSS |
| Infra | Docker, Caddy, Railway |
| CI | GitHub Actions |

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