# Directorio para la beta

`support_directory` y `support_places` son colecciones independientes de las alertas. No caducan como alertas, no elevan estadísticas de incidentes y no generan notificaciones.

Hay 50 distritos de Lima y Callao (43 de Lima y siete del Callao) y 12 de la provincia de Trujillo, incluido Alto Trujillo. Cada ficha tiene las líneas nacionales 105, 116, 106, 100 y 113, con su alcance expresamente indicado. 25 fichas incluyen además un contacto local con fuente institucional; las otras indican en la interfaz que falta contrastar el teléfono municipal. No se inventan números para completar cobertura.

La carga incluye 1.638 lugares permanentes: 1.475 en Lima y Callao y 163 en Trujillo. Se conservan ocho fichas contrastadas previamente y se añaden 1.630 registros del RENIPRESS publicado por SUSALUD el 30 de septiembre de 2026. La cobertura incluye los 43 distritos de Lima, los siete del Callao y los 12 de Trujillo, incluido Alto Trujillo. Incluye hospitales, clínicas, institutos, centros médicos, policlínicos y postas; cada registro importado conserva institución, clasificación, categoría, código IPRESS, ubigeo y fuentes. No se presenta un establecimiento privado como servicio público gratuito.

Fuente: https://www.datosabiertos.gob.pe/dataset/registro-nacional-de-entidades-prestadoras-de-servicios-de-salud-renipress/resource/2d4389ed

Importación reproducible: `python3 scripts/import-support-renipress.py /ruta/RENIPRESS_30-09-2026.csv --reviewed 2026-10-11`. El manifiesto guarda SHA-256 del CSV, fecha de publicación, fecha de consulta y cobertura. Se excluyen registros inactivos, consultorios individuales/laboratorios y fichas sin coordenadas válidas en su provincia o dirección utilizable. No se inventan ubicaciones ni teléfonos. Las coordenadas son de referencia del registro, no una comprobación de campo. Los números ambiguos, múltiples o con anexos se omiten; los teléfonos locales inequívocos se normalizan con el código de área de Lima o Trujillo. No se garantiza atención de urgencias, horario ni disponibilidad.

El directorio no tiene TTL. La exploración de ciudad integra Lima y Callao como una única área metropolitana; el selector conserva los 50 distritos y sus propios ubigeos. `city=lima` representa este directorio conjunto. Las consultas geográficas devuelven todos los lugares del área consultada, sin el antiguo recorte a cien. La web agrupa marcadores cercanos visualmente al alejar el mapa y permite abrir los lugares al acercarse. La app muestra un aviso diferente al de reportes a menos de 500 m del GPS real, con precisión de hasta 150 m. El aviso local usa un canal de importancia normal, requiere notificaciones activadas y se limita a uno por cuenta/dispositivo cada 24 horas. Funciona con el mapa abierto; se suspende el seguimiento al pasar la app a segundo plano. Explorar manualmente otra zona no genera avisos de cercanía. La tarjeta de cercanía sigue visible sin permiso de notificaciones.
El arranque inserta únicamente los registros ausentes mediante `$setOnInsert`. Es idempotente y conserva las ediciones posteriores. Para actualizar información existente debe revisarse su fuente y editarse explícitamente en MongoDB; cambiar un archivo de semilla no sustituye registros existentes.

API pública:

- `GET /v1/support/directory?city=lima|trujillo` (`lima` incluye Callao)
- `GET /v1/support/coverage?lat=-12.061&lng=-77.139`: respuesta pequeña con `covered`, distrito y provincia; no consulta incidentes ni descarga polígonos.
- `GET /v1/support/places?lat=-12.0464&lng=-77.0428&radius_meters=50000`

Web y móvil consumen las mismas fichas. Una cruz blanca en círculo rojo identifica recursos permanentes separados de los reportes comunitarios. Web admite `ubigeo` y `scope=district|city`; móvil también consulta cercanía por coordenadas y radio. Los límites de referencia de 2020 aún incluyen el actual Alto Trujillo dentro de El Porvenir; el directorio conserva el ubigeo actual publicado por SUSALUD.

La app muestra «Tu ciudad aún no tiene cobertura» cuando la ubicación real pertenece a otra provincia. Un fallo de consulta no se interpreta como falta de cobertura. Las líneas nacionales de ayuda permanecen disponibles.
