# Harkai backend

API Go compartida por el frontend web y la app móvil. Autenticación propia, MongoDB, MinIO S3, reportes geográficos, fotos, preferencias y directorio de emergencias. Firebase se limita a verificar identidad Google y entregar notificaciones FCM cuando estén configuradas. La migración de los clientes está documentada en `docs/client-migration.md`.

## Arquitectura

| Directorio | Responsabilidad |
| --- | --- |
| `internal/core` | Modelos, validación y reglas puras; tiempo como argumento |
| `internal/application` | Casos de uso y puertos pequeños |
| `internal/mongostore` | Persistencia, índices y operaciones atómicas |
| `internal/s3store` | Fotos privadas en S3/MinIO |
| `internal/security` | Tokens aleatorios, bcrypt e imágenes normalizadas |
| `internal/httpapi` | API HTTP v1, respuestas, autenticación y CORS |
| `internal/push` | Entrega FCM, reintentos y cola persistente ligada al reporte |
| `internal/config` | Lectura y validación del entorno |
| `internal/runtime` | Composición, arranque y apagado |
| `cmd/api` | Entrada mínima del servidor |
| `cmd/admin` | Creación explícita de una cuenta administrativa nueva |
| `cmd/checkstructure` | Verifica retornos y máximo de 200 líneas por archivo de producción |

Las funciones de producción devuelven resultados o errores. `main` y `ServeHTTP` conservan las firmas sin retorno que exigen Go y net/http. Las firmas de testing y sus callbacks también son obligatorias. El núcleo no conoce MongoDB ni HTTP.

## Desarrollo local

Go 1.27.2, MongoDB 7 y MinIO. `docker compose up --build` inicia MongoDB, MinIO, un bucket privado y la API en `http://localhost:8080`. La base local está separada de producción.

Para ejecución directa, exporta las variables de `.env.example` y ejecuta `go run ./cmd/api`. El servidor no carga archivos `.env` automáticamente. No guardes credenciales en Git.

```sh
gofmt -w cmd internal
go run ./cmd/checkstructure
go test -race ./...
go vet ./...
go build ./cmd/api
```

Las pruebas de integración requieren `TEST_MONGODB_URI` de un MongoDB local aislado y `TEST_S3_ENDPOINT`, `TEST_S3_ACCESS_KEY`, `TEST_S3_SECRET_KEY` de un MinIO local aislado. Crean y eliminan bases aleatorias `harkai_test_*`; no usan `MONGODB_DATABASE`. También crean y eliminan buckets aleatorios `harkai-test-*`. Las credenciales de compose son exclusivas para desarrollo local. Sin esa variable se omiten explícitamente.

## Contrato para los clientes

Consulta `docs/api.md` y `internal/httpapi/openapi.json`. El contrato OpenAPI también se sirve en `/openapi.json`. Los tipos son cadenas estables: `fire`, `crash`, `theft`, `pet`, `emergency`, `place`, `extortion`, `kidnapping`, `violence`, `homicide`. No se aceptan índices numéricos del enum antiguo.

Los usuarios creados por registro público reciben rol `user`. Las fotos se suben antes de crear el reporte; solo su propietario puede adjuntarlas. Los contactos se publican únicamente con consentimiento explícito. Los reportes comienzan como no verificados.

Las alertas por zona están disponibles. El adaptador FCM y su cola persistente están implementados, pero la activación de credenciales y la recepción real siguen pendientes en producción. `/v1/meta` declara las capacidades disponibles. Las reglas de texto, transcripción local de audio y revisión local de imágenes tienen adaptadores reales. Google requiere `GOOGLE_PROJECT_ID`; audio y fotos requieren `MEDIA_WORKER_URL` y `MEDIA_WORKER_KEY`. Recuperación de contraseña, pagos, activación de credenciales push, SSE y migración histórica siguen pendientes; no hay endpoints simulados. El trabajador privado se construye con `media-worker/Dockerfile`; consulta su README para modelos y fuentes.

## Despliegue

Dockerfile multietapa, usuario sin privilegios y puerto 8080. Dokploy aporta `MONGODB_URI`, `MONGODB_DATABASE=harkai_backend` y orígenes CORS exactos. Dominio previsto: `api.harkai.lat`. `/health/live` comprueba el proceso; `/health/ready` comprueba MongoDB y el bucket S3. Detalles en `docs/operations.md`.

## Historial y conversación

GET /v1/archive/incidents consulta el archivo público por zona, tipo, distrito reportado y fechas (after inclusivo, before exclusivo). Incluye reportes vencidos y resueltos; excluye ocultos. No existe un TTL que elimine reportes al vencer la alerta. GET /v1/archive/districts devuelve hasta 200 etiquetas aportadas en esa consulta, sin afirmar límites oficiales. Las fotos aprobadas de reportes archivados siguen accesibles; ocultar el reporte retira ese acceso.

GET/POST /v1/incidents/{id}/comments comparte la misma conversación entre clientes. POST requiere una sesión Go y limita 5/minuto y 30/hora. DELETE permite al autor o al admin retirar texto y nombre públicos. El correo y el ID de cuenta no forman parte del contrato público. La conversación conserva un marcador de comentario retirado.
