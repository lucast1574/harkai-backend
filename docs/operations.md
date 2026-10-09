# Operación de Harkai backend

Organización Dokploy: lucast. Proyecto: harkai. Entorno: production. La instancia MongoDB existente se reutiliza sin eliminar ni migrar colecciones previas. Esta API usa una base independiente `harkai_backend`.

## Variables y despliegue

- PORT: 8080.
- MONGODB_URI: conexión privada a MongoDB; nunca exponer el puerto de MongoDB para alimentar clientes.
- MONGODB_DATABASE: harkai_backend.
- S3_ENDPOINT: harkai-minio:9000 (red privada de Docker), S3_USE_SSL=false.
- S3_ACCESS_KEY y S3_SECRET_KEY: usuario exclusivo, limitado al bucket y prefijo photos; no usar la cuenta root en la API.
- S3_BUCKET: harkai-photos.
- CORS_ORIGINS: lista de orígenes exactos separados por comas, https://panel.harkai.lat, https://harkai.lat y https://www.harkai.lat.

Dokploy construye el Dockerfile desde codex/harkai-rebuild del repositorio privado harkai-backend. El contenedor escucha en 8080 y se ejecuta sin privilegios. El proxy dirige api.harkai.lat a ese puerto. Cloudflare debe apuntar al servidor de Dokploy; HTTPS requiere un certificado válido en origen y en el proxy.

La API está fijada a rangelserver2, donde Dokploy construye su imagen local, con límites de 512 MB y 0.5 CPU. MinIO permanece fijado a rangelserver1, donde existe el HDD. Para mover o replicar la API a otros nodos debe configurarse distribución de imagen mediante registro.

El inicio comprueba conexión e índices antes de aceptar tráfico. La señal SIGTERM inicia apagado con plazo de 15 segundos. La API impone límites de tamaño, timeouts HTTP y CORS exacto. No incluye credenciales en logs ni respuestas. Readiness: `/health/ready`; liveness: `/health/live`.

## Persistencia y copias

Colecciones nuevas: users, sessions, incidents, comments, devices, media, cooldowns, rate_limits y emergency_contacts. Los índices de TTL eliminan sesiones vencidas y registros de límites; el acceso comprueba vencimiento en cada lectura y no depende de la puntualidad de TTL. Los incidentes no se borran al caducar: quedan para historial y revisión. Configura copias y retención de harkai_backend y fotos antes de abrir el producto a usuarios reales.

Los límites HTTP por IP usan RemoteAddr y no confían en cabeceras de identidad del cliente. Detrás de Dokploy la IP puede ser compartida por el proxy; ajustar una cadena de proxies confiables es trabajo posterior y requiere verificar sus redes. Los límites por cuenta y de reportes son persistentes en MongoDB y atómicos. Si se escala el servicio, todos los contenedores deben usar la misma base.

Las fotos sin reporte adjunto permanecen privadas; todavía no hay tarea de eliminación de archivos huérfanos ni borrado de cuenta. No se han migrado datos, usuarios, archivos o contraseñas de Firebase. Las credenciales Firebase no son contraseñas reutilizables de este sistema.

## Almacenamiento de fotos

`deploy/minio.stack.yaml` configura una instancia separada en rangelserver1, fijada al nodo donde existe el HDD. Bind persistente: `/srv/minio-data/harkai/minio`, sobre `/dev/md0` (RAID de discos mecánicos). Límite de memoria 512 MB y CPU 0.5; bucket privado con cuota de 10 GiB. La cuota se aplica según el escaneo de MinIO, no es una partición física de 10 GiB. Sin puertos públicos ni consola expuesta. La API entrega las fotos después de comprobar permisos.

MongoDB solo conserva metadatos. Una caída entre subir S3 y confirmar Mongo puede dejar un objeto huérfano privado: no borrarlo automáticamente ante un resultado incierto de Mongo. Respaldar tanto la base como el HDD; RAID no reemplaza copias. El inicializador de bucket/usuario termina después de configurar almacenamiento. Sus credenciales root quedan solo en el entorno de Dokploy.

## Comprobaciones

Ejecuta las pruebas con una base local aislada y revisa permisos, paginación, geografía, expiración, rotación de sesiones, límites concurrentes y acceso a fotos. Ejecuta el verificador de estructura antes de publicar. En producción verifica HTTPS, `/health/ready`, `/v1/meta`, el contrato OpenAPI y que rutas privadas rechacen llamadas sin token. Evita crear cuentas o reportes de prueba en producción para una comprobación de salud.

## Notificaciones activadas

Firebase queda limitado a Google login y transporte FCM. Go registra dispositivos ligados a una sesión, selecciona reportes por radio/preferencias y conserva la cola en MongoDB. Las alertas no verificadas también pueden avisar; cerrar sesión revoca la elegibilidad.

La credencial privada `FCM_SERVICE_ACCOUNT_BASE64` está solo en el entorno de la API de Dokploy. El remitente `harkai-fcm-sender@harkai-acceso.iam.gserviceaccount.com` tiene únicamente `cloudmessaging.messages.create`. La clave pública VAPID está en `NEXT_PUBLIC_FIREBASE_VAPID_KEY` al construir el panel. Ninguna clave privada pertenece a los clientes ni al repositorio.

Activación del 9 de octubre de 2026: validación FCM `validate_only` correcta y `/v1/meta` con `push_notifications=true`. Esto comprueba autenticación y permiso, no recepción en teléfonos/navegadores. La clave activa `3cc9e27d07764a55b245d6dc984d887a3503f409` vence el 9 de octubre de 2027 a las 14:37:59 UTC: reemplazarla antes de esa fecha, desplegar y validar la nueva antes de retirar la anterior.

## Trabajo pendiente

Recepción push en dispositivos reales, firma Android de publicación, recuperación de contraseña, gestión de territorios institucionales, pagos, actualización en vivo y migración histórica de Firebase. El borrado de cuenta todavía requiere el procedimiento manual publicado; configurar copias y comprobarlo antes de abrir el producto ampliamente.
