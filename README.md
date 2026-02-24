# Kemeny Studio — Prueba Técnica Fullstack

## Contexto

Eres parte de una consultora de software y te asignan a un proyecto existente. Un desarrollador junior construyó esta aplicación de gestión de tareas para un cliente. El cliente quiere:

1. **Una revisión de calidad del código** antes de seguir construyendo sobre él
2. **Un feature nuevo**: clasificación automática de tareas usando IA (LLM)
3. **Infraestructura** para que todo el equipo pueda levantar el proyecto fácilmente con Docker

---

## Stack

- **Backend**: Go (chi router, pgx para PostgreSQL)
- **Frontend**: Next.js 14 (App Router, TypeScript)
- **Base de datos**: PostgreSQL 16
- **IA**: Interface Go para integrar OpenAI o Anthropic

---

## Estructura del Proyecto

```
├── backend/
│   ├── cmd/server/main.go           # Entry point
│   ├── internal/
│   │   ├── handler/tasks.go         # CRUD handlers
│   │   ├── middleware/auth.go       # JWT auth
│   │   ├── model/task.go            # Structs
│   │   ├── db/connection.go         # Pool PostgreSQL
│   │   ├── notification/service.go  # Notificaciones
│   │   └── llm/
│   │       ├── client.go            # Interface LLMClient
│   │       └── mock.go              # Mock para testing
│   └── tests/tasks_test.go
├── frontend/
│   └── src/
│       ├── app/                     # Next.js App Router pages
│       ├── components/              # React components
│       ├── lib/api.ts               # API client
│       └── types/index.ts           # TypeScript types
├── database/
│   └── init.sql                     # Schema + seed data
```

---

## Entregables

### Parte 1 — Code Review (~30 min)

Crea un archivo `REVIEW.md` en la raíz del proyecto:

- Lista los problemas que encuentres, **ordenados por severidad** (crítico → bajo)

### Parte 2 — Fix de Bugs (~20 min)

Arregla al menos los **2 problemas que consideres más críticos** de tu review.

### Parte 3 — Feature: Auto-clasificación con IA (~40 min)

Implementa un endpoint `POST /api/tasks/:id/classify` que:

1. Reciba el ID de una tarea existente
2. Envíe el título y descripción a un LLM para obtener:
   - Tags sugeridos
   - Prioridad recomendada
   - Categoría (bug, feature, improvement, research)
   - Resumen de una línea
3. Guarde la clasificación en la base de datos (tablas `tags` y `task_tags`)
4. Retorne la tarea actualizada con su clasificación


**Implementa un client real** que llame a OpenAI o Anthropic. Si prefieres trabajar solo con el mock, es válido — lo que evaluamos es el diseño de la integración.

### Parte 4 — Docker (~20 min)

Crea Dockerfiles + `docker-compose.yml` para levantar todo con un solo comando:

```bash
docker-compose up
```

### Parte 5 — Arquitectura y Decisiones (~30 min)

Completa el archivo `ARCHITECTURE.md` que ya existe en la raíz del proyecto.

## Cómo Empezar

```bash
# Base de datos
docker run -d --name taskdb \
  -e POSTGRES_PASSWORD=assessment \
  -e POSTGRES_DB=taskmanager \
  -p 5432:5432 \
  postgres:16-alpine

# Seed
PGPASSWORD=assessment psql -h localhost -U postgres -d taskmanager -f database/init.sql

# Backend
cd backend
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=assessment DB_NAME=taskmanager
export JWT_SECRET=dev-secret
go run cmd/server/main.go

# Frontend (otra terminal)
cd frontend
npm install
NEXT_PUBLIC_API_URL=http://localhost:8080 npm run dev
```

### API de ejemplo

```bash
# Login (acepta cualquier password para usuarios del seed)
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "carlos@kemeny.studio", "password": "password123"}'

# Listar tareas (usar el token del login)
curl http://localhost:8080/api/tasks \
  -H "Authorization: Bearer <TOKEN>"

# Obtener tarea con detalle
curl http://localhost:8080/api/tasks/11111111-1111-1111-1111-111111111111 \
  -H "Authorization: Bearer <TOKEN>"
```

---

## Reglas

- **Puedes usar cualquier herramienta de IA para desarrollar** (Claude Code, Copilot, ChatGPT, Cursor, etc.). De hecho, lo alentamos.
- **Si quieres compartir tu historial de chat con el modelo que usaste, suma bastante.** Nos ayuda a entender tu proceso de pensamiento y cómo interactúas con las herramientas. No es obligatorio, pero es un plus.
- **No hace falta que sea perfecto.** Buen criterio > código impecable. Si algo te llevaría demasiado tiempo, documenta qué harías en el `ARCHITECTURE.md`.
- **Tiempo estimado: ~2.5 horas.**.

---

## Entrega

1. Haz un **fork** de este repositorio a tu cuenta de GitHub
2. Trabaja sobre tu fork y asegúrate de que todos tus cambios estén commiteados
3. Verifica que existan: `REVIEW.md`, `ARCHITECTURE.md`, `docker-compose.yml`
4. Cuando hayas terminado, envíanos la confirmación con el link a tu fork para que podamos revisarlo

---

## Usuarios del Seed

| Email | Nombre | Rol |
|-------|--------|-----|
| carlos@kemeny.studio | Carlos Méndez | admin |
| lucia@kemeny.studio | Lucía Fernández | member |
| mateo@kemeny.studio | Mateo Ruiz | member |
| valentina@kemeny.studio | Valentina López | member |

Password para todos: `password123`