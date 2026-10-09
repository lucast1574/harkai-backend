# Operación de Harkai backend

Organización Dokploy: lucast. Proyecto: harkai. Entorno: production. La instancia MongoDB existente se reutiliza sin eliminar ni migrar colecciones previas. Esta API usa una base independiente `harkai_backend`.

## Variables y despliegue

- PORT: 8080.
- MONGODB_URI: conexión privada a MongoDB; nunca exponer el puerto de MongoDB para alimentar clientes.
- MONGODB_DATABASE: harkai_backend.
- S3_ENDPOINT: harkai-minio:9000 (red privada de Docker), S3_USE_SSL=false.
- S3_ACCESS_KEY y S3_SECRET_KEY: usuario exclusivo, limitado al bucket y prefijo photos; no usar la cuenta root en la API.
- S3_BUCKET: harkai-photos.
- CORS_ORIGINS: lista de orígenes exactos separados por comas, inicialmente https://harkai.lat y https://www.harkai.lat.

Dokploy construye el Dockerfile desde main del repositorio privado harkai-backend. El contenedor escucha en 8080 y se ejecuta sin privilegios. El proxy dirige api.harkai.lat a ese puerto. Cloudflare debe apuntar al servidor de Dokploy; HTTPS requiere un certificado válido en origen y en el proxy.

El inicio comprueba conexión e índices antes de aceptar tráfico. La señal SIGTERM inicia apagado con plazo de 15 segundos. La API impone límites de tamaño, timeouts HTTP y CORS exacto. No incluye credenciales en logs ni respuestas. Readiness: `/health/ready`; liveness: `/health/live`.

## Persistencia y copias

Colecciones nuevas: users, sessions, incidents, media, cooldowns, rate_limits y emergency_contacts. Los índices de TTL eliminan sesiones vencidas y registros de límites; el acceso comprueba vencimiento en cada lectura y no depende de la puntualidad de TTL. Los incidentes no se borran al caducar: quedan para historial y revisión. Configura copias y retención de harkai_backend y fotos antes de abrir el producto a usuarios reales.

Los límites HTTP por IP usan RemoteAddr y no confían en cabeceras de identidad del cliente. Detrás de Dokploy la IP puede ser compartida por el proxy; ajustar una cadena de proxies confiables es trabajo posterior y requiere verificar sus redes. Los límites por cuenta y de reportes son persistentes en MongoDB y atómicos. Si se escala el servicio, todos los contenedores deben usar la misma base.

Las fotos sin reporte adjunto permanecen privadas; todavía no hay tarea de eliminación de archivos huérfanos ni borrado de cuenta. No se han migrado datos, usuarios, archivos o contraseñas de Firebase. Las credenciales Firebase no son contraseñas reutilizables de este sistema.

## Almacenamiento de fotos

`deploy/minio.stack.yaml` configura una instancia separada en rangelserver1, fijada al nodo donde existe el HDD. Bind persistente: `/srv/minio-data/harkai/minio`, sobre `/dev/md0` (RAID de discos mecánicos). Límite de memoria 512 MB y CPU 0.5; bucket privado con cuota de 10 GiB. La cuota se aplica según el escaneo de MinIO, no es una partición física de 10 GiB. Sin puertos públicos ni consola expuesta. La API entrega las fotos después de comprobar permisos.

MongoDB solo conserva metadatos. Una caída entre subir S3 y confirmar Mongo puede dejar un objeto huérfano privado: no borrarlo automáticamente ante un resultado incierto de Mongo. Respaldar tanto la base como el HDD; RAID no reemplaza copias. El inicializador de bucket/usuario termina después de configurar almacenamiento. Sus credenciales root quedan solo en el entorno de Dokploy.

## Comprobaciones

Ejecuta las pruebas con una base local aislada y revisa permisos, paginación, geografía, expiración, rotación de sesiones, límites concurrentes y acceso a fotos. Ejecuta el verificador de estructura antes de publicar. En producción verifica HTTPS, `/health/ready`, `/v1/meta`, el contrato OpenAPI y que rutas privadas rechacen llamadas sin token. Evita crear cuentas o reportes de prueba en producción para una comprobación de salud.

## Siguientes integraciones

Los clientes deben migrarse explícitamente al contrato v1. Quedan fuera de esta base: Google OAuth, recuperación de contraseña, gestión de cuentas empresariales/territorios, IA y audio, pagos reales, APNs/Web Push, actualización en vivo y migración Firebase. Para eliminar Firebase también de las alertas Android se debe elegir un transporte push independiente y comprobar su funcionamiento en los dispositivos objetivo. Las alertas consultables actuales no ofrecen entrega con la app cerrada.
