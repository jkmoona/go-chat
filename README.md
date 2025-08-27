# go-chat

A simple real-time chat app built with **Go (Gin + WebSockets)**, **Postgres**, and a **Vue client**.  

## Features

- JWT auth (register, login, refresh)
- Rooms & private messaging via WebSockets
- Dockerized backend with Postgres
- Vue 3 + Tailwind frontend
- **Live demo**: [tempchatgo.netlify.app](https://tempchatgo.netlify.app)

## Setup

### Backend

```sh
cd server
cp .env.example .env
docker compose up --build
```

### Client

```sh
cd client
cp .env.example .env
npm install
npm run dev
```

## Cleanup

To stop and remove all containers, networks, and volumes created by this project:

```sh
cd server
docker compose down -v
```

## License

This project is licensed under the [MIT License](LICENSE).
