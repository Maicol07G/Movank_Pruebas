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

## Variables
Copiar `.env.example` como `.env` y ajustar si es necesario.

## Decisiones
Consultar `docs/DECISIONS.md` y `docs/API.md`.

> Nota: esta base está diseñada para estudiar y completar la prueba de forma entendible. Antes de entregar, ejecutar las pruebas y revisar cada decisión.
