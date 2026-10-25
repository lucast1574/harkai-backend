# Directorio para la beta

`support_directory` y `support_places` son colecciones independientes de las alertas. No caducan como alertas, no elevan estadísticas de incidentes y no generan notificaciones.

Hay 50 distritos de Lima y Callao (43 de Lima y siete del Callao) y 12 de la provincia de Trujillo, incluido Alto Trujillo. Cada ficha tiene las líneas nacionales 105, 116, 106, 100 y 113, con su alcance expresamente indicado. 25 fichas incluyen además un contacto local con fuente institucional; las otras indican en la interfaz que falta contrastar el teléfono municipal. No se inventan números para completar cobertura.

La carga nacional incluye 11.662 establecimientos RENIPRESS con coordenadas y dirección utilizables, en los 25 departamentos del Perú, además de las ocho fichas contrastadas previamente. La publicación de SUSALUD es del 30 de septiembre de 2026. Incluye hospitales, clínicas, institutos, centros médicos, policlínicos y postas; cada registro importado conserva departamento, provincia, institución, clasificación, categoría, código IPRESS, ubigeo y fuentes. No se presenta un establecimiento privado como servicio público gratuito.

El directorio telefónico conserva los contactos municipales contrastados de Lima, Callao y Trujillo. Para los demás distritos del catálogo geográfico se publican únicamente líneas nacionales verificadas; no se inventan teléfonos municipales. La web permite seleccionar departamento, provincia y distrito, con accesos rápidos a Lima y Callao y Trujillo. Los límites usan la cartografía SENACE/INEI de referencia 2020; no se inventan polígonos de distritos posteriores.

Fuente: https://www.datosabiertos.gob.pe/dataset/registro-nacional-de-entidades-prestadoras-de-servicios-de-salud-renipress/resource/2d4389ed

Importación reproducible: `python3 scripts/import-support-renipress.py /ruta/RENIPRESS_30-09-2026.csv --reviewed 2026-10-24`. El manifiesto guarda SHA-256 del CSV, fecha de publicación, fecha de consulta y cobertura. Se excluyen registros inactivos, consultorios individuales/laboratorios y fichas sin coordenadas válidas dentro del rango geográfico del Perú o dirección utilizable. No se inventan ubicaciones ni teléfonos. Las coordenadas son de referencia del registro, no una comprobación de campo. Los números ambiguos o con anexos se omiten; los teléfonos múltiples separados explícitamente se conservan como contactos independientes; los teléfonos locales inequívocos se normalizan con el código de área de Lima o Trujillo. No se garantiza atención de urgencias, horario ni disponibilidad.

El directorio no tiene TTL. La exploración de ciudad integra Lima y Callao como una única área metropolitana; el selector conserva los 50 distritos y sus propios ubigeos. `city=lima` representa este directorio conjunto. Las consultas geográficas devuelven todos los lugares del área consultada, sin el antiguo recorte a cien. La web agrupa marcadores cercanos visualmente al alejar el mapa y permite abrir los lugares al acercarse. La app muestra un aviso diferente al de reportes a menos de 500 m del GPS real, con precisión de hasta 150 m. El aviso local usa un canal de importancia normal, requiere notificaciones activadas y se limita a uno por cuenta/dispositivo cada 24 horas. Funciona con el mapa abierto; se suspende el seguimiento al pasar la app a segundo plano. Explorar manualmente otra zona no genera avisos de cercanía. La tarjeta de cercanía sigue visible sin permiso de notificaciones.
El arranque inserta únicamente los registros ausentes mediante `$setOnInsert`. Es idempotente y conserva las ediciones posteriores. Las consultas completan teléfonos y fotos ausentes con el catálogo RENIPRESS contrastado, sin sobrescribir contactos o fotos editados en MongoDB. Los demás campos existentes solo se actualizan explícitamente en MongoDB.

API pública:

- `GET /v1/support/directory?city=lima|trujillo` (`lima` incluye Callao)
- `GET /v1/support/coverage?lat=-12.061&lng=-77.139`: respuesta pequeña con `covered`, distrito y provincia; no consulta incidentes ni descarga polígonos.
- `GET /v1/support/places?lat=-12.0464&lng=-77.0428&radius_meters=50000`

Web y móvil consumen las mismas fichas. Una cruz blanca en círculo rojo identifica recursos permanentes separados de los reportes comunitarios. Web admite `ubigeo` y `scope=district|city`; móvil también consulta cercanía por coordenadas y radio. Los límites de referencia de 2020 aún incluyen el actual Alto Trujillo dentro de El Porvenir; el directorio conserva el ubigeo actual publicado por SUSALUD.

La app muestra «Tu ciudad aún no tiene cobertura» cuando la ubicación real pertenece a otra provincia. Un fallo de consulta no se interpreta como falta de cobertura. Las líneas nacionales de ayuda permanecen disponibles.

## Ficha móvil de centros médicos

La ficha muestra únicamente nombre, foto disponible y contactos que abren el marcador del equipo mediante `tel:`. Conserva las fuentes y clasificación en los datos, sin mostrarlas como bloques de texto. El contrato añade `phones` y `photo_urls`, manteniendo `phone` para clientes anteriores.

El CSV de septiembre de 2026 conserva `IMAGEN_1`, `IMAGEN_2` e `IMAGEN_3` y normaliza teléfonos internacionales como `(51) 1 - 2190000`. La revisión de 2026-10-23 recupera fotos para 1.628 de los 1.630 registros importados y teléfonos para 1.465. Los enlaces pertenecen al servicio legado de SUSALUD; no se garantiza su disponibilidad. Android permite su HTTP únicamente para `app20.susalud.gob.pe`; los demás dominios mantienen HTTPS. Una foto fallida se oculta y un centro sin teléfono indica que no está disponible.
