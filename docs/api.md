# Contrato de la API Harkai

Base de producción: `https://api.harkai.lat`. Versión HTTP: `/v1`. JSON en snake_case y fechas UTC RFC3339. El frontend y Flutter deben consumir el mismo contrato. El formato heredado Firestore no se acepta automáticamente.

## Autenticación

| Método | Ruta | Resultado |
| --- | --- | --- |
| POST | `/v1/auth/register` | Crea usuario con rol user y sesión; 201 |
| POST | `/v1/auth/login` | Inicia sesión; 200 |
| POST | `/v1/auth/refresh` | Consume refresh_token y rota ambos tokens; 200 |
| POST | `/v1/auth/logout` | Revoca la sesión actual y ambos tokens; 204 |
| GET | `/v1/me` | Perfil de la sesión |
| PATCH | `/v1/me` | Actualiza name |
| PUT | `/v1/me/preferences` | Sustituye preferencias validadas |

Registro: `{"email":"usuario@example.com","name":"Usuario","password":"una-clave-larga-y-unica"}`. Login acepta email y password. Contraseña de 12 a 72 bytes al registrarse. No se aceptan campos como role ni password_hash.

La respuesta de autenticación contiene user, access_token, refresh_token, token_type y expires_in. Usa `Authorization: Bearer <access_token>` en rutas privadas. Access token: 15 minutos; sesión: 30 días desde su creación. El refresh no extiende ese plazo. Al rotar, el access anterior deja de funcionar y el refresh anterior no se puede reutilizar. Los tokens se almacenan como hashes SHA-256 en MongoDB; las contraseñas como bcrypt con coste 12.

El móvil debe guardar refresh tokens en almacenamiento seguro del sistema operativo. El frontend debe mantener el access en memoria y usar una capa de servidor con cookie HttpOnly para custodiar el refresh; no almacenarlo en localStorage. La API v1 devuelve tokens JSON y no establece cookies por sí misma.

Preferencias completas: `{"notifications":"critical","radius_meters":500,"language":"es","theme":"system"}`. Notificaciones: all, critical, none. Radio entre 100 y 10000 metros; idioma es/en; tema system/light/dark. Esas preferencias se aplican en `/v1/alerts`.

## Reportes

| Método | Ruta | Acceso |
| --- | --- | --- |
| GET | `/v1/incidents` | Público, consulta geográfica obligatoria |
| POST | `/v1/incidents` | Usuario autenticado |
| GET | `/v1/incidents/{id}` | Público salvo reportes ocultos |
| PATCH | `/v1/incidents/{id}/status` | Propietario o administrador |
| GET | `/v1/me/incidents` | Propios, incluyendo inactivos |
| GET | `/v1/admin/incidents` | Administrador, incluyendo inactivos dentro de una zona |
| GET | `/v1/alerts` | Usuario, radio y tipos según preferencias |

```json
{
  "type": "fire",
  "description": "Se observa humo y fuego en el edificio",
  "latitude": -8.1,
  "longitude": -79.0,
  "media_id": "id_opcional_de_foto_propia",
  "share_contact": false
}
```

La respuesta añade id, status, verified, created_at, updated_at y expires_at. Nunca revela user_id. description tiene entre 5 y 2000 caracteres. district/city/country son etiquetas aportadas por el usuario, no direcciones verificadas por el servidor. Los permisos del GPS y la distancia entre el dispositivo y el punto pertenecen al cliente; las coordenadas reportadas no son prueba de presencia.

Mascotas, lugares y eventos requieren media_id y contact_info. share_contact debe ser true para exponer ese contacto. Eventos requieren ends_at futuro y dentro de 90 días. Caducidad de incendios, accidentes, robos y emergencias: 3 horas; mascotas: 24 horas; lugares: sin caducidad automática; eventos: ends_at. La consulta pública excluye inactivos y caducados independientemente de si hay clientes abiertos. El detalle mantiene el historial de reportes resueltos/caducados; los ocultos solo son visibles para propietario/admin.

Cambiar estado acepta `{"status":"resolved"}` o `{"status":"hidden"}`. No existe reactivación pública. El estado no convierte el reporte en verificado.

Consulta de ejemplo: `/v1/incidents?lat=-8.1&lng=-79&radius_meters=5000&type=fire&limit=50`. Radio: 100 a 100000 metros; limit: 1 a 100. Páginas ordenadas por id descendente, no distancia. Respuesta: `{"items":[],"next_cursor":"..."}`; envía next_cursor como cursor y conserva los filtros. after opcional filtra created_at mayor a una fecha RFC3339. Conserva el mismo after al recorrer páginas; avanza el punto de consulta después de completar el recorrido. Con filtro crítico una página puede estar vacía y todavía tener next_cursor.

Límites persistidos por cuenta: 20 intentos de creación por hora y un reporte por tipo por hora salvo lugares/eventos. El cooldown se reserva antes de insertar: un fallo incierto de almacenamiento conserva la reserva para evitar duplicados. Las actualizaciones de estado y las sesiones usan escrituras condicionales; funcionan con varias réplicas.

## Fotos

POST `/v1/media` recibe bytes JPEG/PNG con Content-Type image/jpeg o image/png y token Bearer. Máximo 5 MiB, 4096 por dimensión y 12 millones de píxeles. Se decodifica y reescribe en JPEG para descartar metadatos y contenido adjunto. Devuelve id, content_type, size y created_at; adjunta ese id al reporte.

GET `/v1/media/{id}` devuelve JPEG. Antes de publicar solo puede verlo el propietario/admin. Al adjuntarse a un reporte activo y vigente se vuelve público; se retira del acceso público si el reporte se oculta, resuelve o caduca. Máximo 20 intentos de subida por usuario/hora. Las imágenes se guardan en MinIO S3 privado, en el HDD del servidor. MongoDB solo guarda metadatos. Los puertos MediaStore y PhotoStore mantienen el contrato independiente del almacenamiento.

## Directorio y administración

GET `/v1/emergency-contacts?country=PE&city=trujillo` devuelve números configurados para esa región. Si no hay registro, devuelve 404; no inventa un número global de emergencia. PUT `/v1/admin/emergency-contacts` permite que un administrador mantenga country, city y numbers con claves police/firefighters/medical/municipal. Los valores deben ser números de teléfono revisados para la región.

Para crear un administrador nuevo utiliza `cmd/admin` con ADMIN_EMAIL, ADMIN_NAME y ADMIN_PASSWORD en el entorno. La herramienta rechaza cuentas existentes, no cambia sus credenciales ni asigna privilegios mediante registro público. Ninguna cuenta administrativa se crea automáticamente al arrancar.

## Errores y capacidades

Errores JSON: `{"error":{"code":"invalid_input","message":"..."}}`. Estados: 400 validación, 401 autenticación, 403 permisos, 404 recurso, 409 duplicado, 429 límite, 500 fallo interno, 503 no disponible. Las rutas JSON rechazan campos desconocidos y múltiples documentos en el mismo body.

GET `/v1/meta` informa tipos y capacidades. GET `/health/live` y `/health/ready` son públicos. Push, pagos e IA aparecen como false hasta existir una implementación real.
