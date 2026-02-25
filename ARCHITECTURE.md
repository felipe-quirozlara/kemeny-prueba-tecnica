# Arquitectura y Decisiones

## 1. Code Review — Resumen Ejecutivo

¿Cuáles son los 2-3 problemas más graves y por qué los priorizaste?

Lo primero que modifiqué fue el proyecto para que pueda iniciar, los problemas con el seed no permitian iniciar el proyecto. Además en front end sin un token valido configurado a través de localStorage no permitiría consumir el dashboard o la página de task. Se modificó el init.sql y se modificó el Front end para agregar un login page para poder utilizar la herramientas.
El error del token con exp como string no permitia validar el token y poder recibir respuesta del endpoint de task o para el dashboard, lo que impide el uso, un error grave y priorizado.
Endpoint ListTask tenía un problema de optimización utilizando dos bucle para obtener información, se modificó para optimizar la llamadas y consultas.

## 2. Integración de IA

- ¿Qué modelo/provider elegiste y por qué?
- ¿Cómo diseñaste el prompt? ¿Qué iteraciones hiciste?
- ¿Cómo manejas los casos donde el LLM falla, tarda, o retorna datos inválidos?
- ¿Cómo manejarías el costo en producción? (caching, rate limiting, batch processing)
- Si tuvieras que clasificar 10,000 tareas existentes, ¿cómo lo harías?

- Provider: el código soporta OpenAI y Anthropic y selecciona automáticamente según las variables de entorno (ver `backend/cmd/server/main.go`). Si no hay keys se usa `MockClient` (ver `backend/internal/llm/mock.go`). Por defecto el proyecto prefiere OpenAI si detecta `OPENAI_API_KEY`.

- Prompt: la implementación entrega al LLM una instrucción que pide "Reply with JSON only" con la forma esperada (ver `backend/internal/llm/openai.go:buildPrompt` y `backend/internal/llm/anthropic.go`). El formato fuerza un objeto JSON con `tags, priority, category, summary` para facilitar el parsing.

- Manejo de fallos y latencia: las llamadas al LLM se hacen con timeout (`context.WithTimeout(..., 10*time.Second)` en `backend/internal/handler/tasks.go:ClassifyTask`). Errores del SDK se propagan y se responden como 502 al cliente. Además hay:
  - Fallback a `MockClient` cuando no hay claves (inicio en `backend/cmd/server/main.go`).
  - Parsing robusto: se extrae el primer objeto JSON del texto de respuesta (`parseJSONFromString` en `backend/internal/llm/openai.go`) y se validan/normalizan valores en el handler (prioridades/categorías válidas, truncado de summary, sanitización de tags en `backend/internal/handler/tasks.go`).

- Control de costos en producción (recomendado): cachear respuestas de clasificación (Redis) para tareas inmutables, agrupar clasificaciones (batch) y procesarlas en segundo plano, aplicar rate limiting por tenant/usuario, y preferir modelos más económicos para procesamientos en lote; guardar resultados persistentes para evitar reconsultas.

- Clasificar 10,000 tareas (estrategia práctica): no hacerlo sincrónicamente desde la API pública. En su lugar:
  1) Poner tareas en una cola (ej. Redis/RQ o RabbitMQ).
  2) Consumidores workers que procesan en paralelo con backoff y límite de concurrencia según cuota LLM.
  3) Usar caché y deduplicación (hash de title+description) para evitar llamadas redundantes.
  4) Usar modelos más baratos para la mayoría y escalar a modelos mejores solo cuando sea necesario.
  5) Registrar métricas y fallas para reintentos manuales/automáticos.


## 3. Docker y Orquestación

¿Qué decisiones tomaste? (multi-stage builds, networking, volumes, health checks, etc.)

- Multi-stage builds: el `backend/Dockerfile` usa multi-stage (builder -> runtime) para producir un binario estático (`CGO_ENABLED=0 ... go build`) y así reducir la imagen final.

- Compose y networking: `docker-compose.yml` orquesta tres servicios (`db`, `backend`, `frontend`). El frontend usa `NEXT_PUBLIC_API_URL: http://backend:8080` cuando corre dentro de compose, y `backend` depende de `db` con `condition: service_healthy`.

- Volúmenes y seed: la base de datos monta `./database/init.sql` en `/docker-entrypoint-initdb.d/init.sql` para inicializar datos (ver `docker-compose.yml`), y persiste datos en `pgdata`.

- Healthchecks y restart: el servicio `db` tiene healthcheck (`pg_isready`) y tanto backend como frontend usan `restart: unless-stopped`.

- Mejoras recomendadas: añadir healthchecks para `backend` y `frontend`, usar secrets para variables sensibles (no hardcodear `JWT_SECRET`), definir una red específica, usar imágenes más seguras (distroless) y añadir readiness/liveness probes si se despliega en k8s.


## 4. Arquitectura para Producción

Si este proyecto fuera a producción con 10,000 usuarios concurrentes:

- ¿Qué cambiarías en la arquitectura?
- ¿Qué servicios agregarías? (cache, queue, CDN, etc.)
- ¿Cómo manejarías el deployment?
- Incluye un diagrama si te parece útil (ASCII, Mermaid, o imagen)

- Cambios clave:
  - Separar responsabilidades: desplegar `frontend` en CDN/hosting (Vercel, S3+CloudFront), `backend` en un clúster (Kubernetes/ECS) detrás de un load balancer.
  - Base de datos gestionada (Postgres RDS/Cloud SQL) con réplicas de lectura y `pgbouncer` para connection pooling.

- Servicios a agregar:
  1) Cache: Redis (cache de consultas, sesiones, respuestas LLM).
 2) Queue: Redis/RabbitMQ (procesamiento asíncrono de clasificación LLM, envío de notificaciones).
 3) CDN: para servir assets y la app frontend.
 4) Observabilidad: Prometheus + Grafana, y agregador de logs (ELK / Loki).
 5) Secrets manager: Vault/Secret Manager para llaves LLM y JWT.
 6) Sistema de rate-limiting / API gateway (Traefik, Kong, o API Gateway cloud).

- Deployment:
  - CI/CD: pipeline que construye imágenes, ejecuta tests, publica imágenes a registry y despliega mediante GitOps o pipelines declarativos; usar despliegues canary/blue-green para minimizar riesgos.
  - Escalado automático: HPA en k8s o autoscaling groups en ECS.

## 5. Trade-offs

¿Qué decisiones tomaste donde había más de una opción válida?

- Sincronía vs asincronía para LLM: la implementación actual hace una llamada síncrona con timeout (sencilla para UX inmediato), pero eso impacta latencia. Alternativa: procesar en background y notificar cuando esté listo (mejor escalabilidad).

- Dependencia de SDKs oficiales (OpenAI/Anthropic) vs HTTP directo: usar SDKs simplifica la integración (manejo de tipos, abstracciones) pero añade dependencias; se mitigó con una `MockClient` para tests.

- SQL JOINs en `ListTasks`: se optó por una consulta con LEFT JOIN para reducir roundtrips, sacrificando cierta complejidad en el code-path de ensamblado de tags (pero mejora performance frente a N+1 queries).

- Auth simple (JWT con secret en env + localStorage en frontend): rápido para evaluación y local dev; en producción preferiría cookies httpOnly o un sistema de sessions más robusto y rotación de secretos.


## 6. Qué Mejorarías con Más Tiempo

Sé específico y prioriza.

- Alta prioridad:
  1) Mover la clasificación LLM a jobs asíncronos (cola + workers) para evitar bloqueo de la API.
 2) Añadir pruebas E2E e integración continua (CI) que corran los tests del backend (`backend/tests`) y del frontend.

- Media prioridad:
 4) Observabilidad: métricas, tracing distribuido, agregación de logs.
 5) Graceful shutdown y migraciones automatizadas de DB.



## 7. Uso de IA (como herramienta de desarrollo)

¿Usaste IA para desarrollar? ¿Para qué? ¿Modificaste lo que te sugirió?

- Sí: hay evidencia en `USE_AI.md` sobre uso de AI para generar el helper de responses y para planear la integración con OpenAI/Anthropic. Se usaron chats y prompts guardados (links en `USE_AI.md`).
- Se modificaron y adaptaron las sugerencias de IA: las respuestas generadas por los LLM se normalizan y validan en `backend/internal/llm/*` y en el handler (`backend/internal/handler/tasks.go`)
