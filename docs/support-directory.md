# Directorio para la beta

`support_directory` y `support_places` son colecciones independientes de las alertas. No caducan como alertas, no elevan estadísticas de incidentes y no generan notificaciones.

Hay 43 distritos de Lima Metropolitana y 12 de la provincia de Trujillo, incluido Alto Trujillo. Cada ficha tiene las líneas nacionales 105, 116, 106, 100 y 113, con su alcance expresamente indicado. 25 fichas incluyen además un contacto local con fuente institucional; las otras indican en la interfaz que falta contrastar el teléfono municipal. No se inventan números para completar cobertura.

La primera carga incluye ocho hospitales y centros de salud. Es una selección inicial, no un inventario completo de establecimientos por distrito. Se almacenan la fuente institucional, la fuente de coordenadas y la fecha de consulta. Algunos documentos institucionales son anteriores a la beta; esa fecha expresa cuándo se consultó la fuente, no una llamada de comprobación. Se omite el teléfono cuando no se encontró una fuente institucional que lo respalde. Los contactos de establecimientos se presentan como contactos institucionales, sin atribuirles atención de urgencias o disponibilidad garantizada.

El arranque inserta únicamente los registros ausentes mediante `$setOnInsert`. Es idempotente y conserva las ediciones posteriores. Para actualizar información existente debe revisarse su fuente y editarse explícitamente en MongoDB; cambiar un archivo de semilla no sustituye registros existentes.

API pública:

- `GET /v1/support/directory?city=lima|trujillo`
- `GET /v1/support/places?lat=-12.0464&lng=-77.0428&radius_meters=50000`

Web y móvil consumen las mismas fichas. Los marcadores azules representan recursos permanentes y están separados de los reportes comunitarios. La selección de distrito filtra teléfonos; el mapa de recursos consulta cercanía y radio.
