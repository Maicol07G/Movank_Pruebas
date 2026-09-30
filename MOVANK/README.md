# MOVANK — Prueba Técnica Full Stack

Proyecto de referencia para la prueba técnica MOVANK.

## Alcance
Full Stack: backend Go + PostgreSQL + Dragonfly + frontend SvelteKit/PWA.

## Requisitos
- Go 1.27+
- Node.js 24+
- Docker Desktop con Docker Compose

## Levantar infraestructura
```powershell
docker compose up -d
docker compose ps
```

## Backend
```powershell
cd backend
go mod tidy
go run ./cmd/server
```

API: http://localhost:8080

## Frontend
```powershell
cd frontend
npm install
npm run dev -- --host
```

