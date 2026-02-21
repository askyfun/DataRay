# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DataRay is a drag-and-drop BI visualization analysis platform (MVP). Monorepo with frontend (React/TypeScript) and backend (Go).

## Architecture

- **Frontend**: React 18 + TypeScript + Ant Design + ECharts + Zustand + @dnd-kit + Vite
- **Backend**: Go 1.26 + Gin + bun ORM + PostgreSQL
- **Deployment**: Docker + docker-compose

Backend routes registered in `backend/cmd/main.go`. Database models in `backend/internal/model/`. Frontend pages in `frontend/src/pages/`. Data source drivers in `backend/internal/datasource/`.

## Common Commands

```bash
# Frontend
cd frontend && npm install
npm run dev          # dev server on port 3000
npm run build        # production build

# Backend
cd backend && go mod download
go run ./cmd/main.go -f etc/config.toml  # dev server on port 8080
go test ./...                            # run all tests
go test -v ./cmd -run TestName           # run single test

# Docker
make docker-up    # start all services
make docker-down  # stop services
make dev          # local dev with hot reload (air)
```

## Key Constraints

- **JSON field naming**: Use snake_case in API (e.g., `table_name`, not `tableName`)
- **TypeScript**: Use `unknown` instead of `any`
- **Go error handling**: Never use empty catch blocks

## Documentation

Detailed docs in `docs/` directory: setup.md, architecture.md, coding-style.md, api-spec.md, todo.md. See `AGENTS.md` for project constraints and resources.
