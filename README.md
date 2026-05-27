# Telegram Family Tree Mini App

MVP Telegram Web App for building family trees: `initData` auth, trees, people, `parent`/`spouse` relations, and interactive SVG visualization with drag/zoom.

## Stack

- Frontend: React, TypeScript, Vite, TailwindCSS, Telegram Web Apps SDK
- Backend: Go 1.24, Gin, PostgreSQL, goose migrations
- Infra: Docker Compose

## Quick Start

1. Copy env file:

```bash
cp .env.example .env
```

2. Set `TELEGRAM_BOT_TOKEN` in `.env` for real Telegram validation.

3. Run the app:

```bash
docker compose up --build
```

Services:

- Frontend: http://localhost:3000
- Backend: http://localhost:8080
- Health check: http://localhost:8080/health

The backend applies SQL migrations on container start.

## Telegram Setup

1. Create a bot with BotFather.
2. Add the frontend URL as the Web App URL.
3. Put the bot token into `.env`:

```env
TELEGRAM_BOT_TOKEN=123456:your-token
TELEGRAM_WEB_APP_URL=https://your-public-https-url.example
```

For local browser development without Telegram, the frontend sends demo `initData`. This works only when `TELEGRAM_BOT_TOKEN` is empty.

Start the bot after setting `TELEGRAM_BOT_TOKEN` and `TELEGRAM_WEB_APP_URL`:

```bash
docker-compose --profile bot up -d --build
```

## Development

Backend:

```bash
cd backend
go mod tidy
go test ./...
go run ./cmd/server
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

Use `VITE_API_URL=http://localhost:8080` for local frontend requests.

## API

Auth:

- `POST /api/auth/telegram` with `{ "init_data": "..." }`

Trees:

- `GET /api/trees`
- `POST /api/trees`
- `GET /api/trees/:id`

Persons:

- `POST /api/persons`
- `PATCH /api/persons/:id`
- `DELETE /api/persons/:id`

Relations:

- `POST /api/relations`
- `DELETE /api/relations/:id`

Supported relation types: `parent`, `spouse`.

## Seed

After the stack is running:

```bash
make seed
```

The seed command creates a demo user, tree, people, and relations.
