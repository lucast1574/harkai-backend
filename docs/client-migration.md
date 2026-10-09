# Clientes web y móvil

Los clientes nuevos consultan la API Go v1. Firestore, Storage, Gemini, Google Maps, pagos y eventos salen del flujo activo. Firebase queda como adaptador de identidad Google y entrega FCM; los permisos siempre vienen de Go.

El panel separa rutas, componentes visuales, funcionalidades y adaptadores. Las sesiones usan cookies HttpOnly con renovación protegida por origen. El proxy limita cuerpos incluso sin Content-Length. El móvil usa almacenamiento seguro y renovación compartida para solicitudes concurrentes.

Los reportes incluyen descripción, ubicación/evidencia y revisión explícita. Las reglas sugieren categorías y recomendaciones. El audio se transcribe localmente, se convierte en texto editable y pasa por las mismas reglas. Una transcripción no publica ni confirma reportes.

Las fotos se normalizan sin metadatos y se revisan antes de escribir en MinIO. Un servicio ausente o una decisión incompleta bloquea la carga. Los medios anteriores sin aprobación registrada tampoco se sirven públicamente. OpenCV procesa la imagen; ONNX Runtime ejecuta el detector local. Este filtro requiere validación con muestras y no garantiza detección absoluta.

Roles user/gov/admin se autorizan en Go. Gov puede consultar reportes y exportar la muestra cargada; el panel identifica resultados parciales. Admin administra roles y moderación. Ganancias indica pagos desactivados.

## Pendiente antes de cerrar la migración

- Probar el acceso real con Google en los dispositivos objetivo; Android/web ya usan harkai-acceso.
- Comprobar la recepción real FCM en Android y navegador. El registro privado de dispositivos, la cola durable y las credenciales están conectados; producción declara push_notifications=true.
- Revisar con usuarios el trabajador de medios, API, panel y landing ya desplegados en Dokploy.
- Validar STT con voces peruanas y el filtro de fotos con un conjunto de evaluación.
- Configurar contactos institucionales y una fuente editorial de noticias. Actualidad actualmente muestra reportes comunitarios.
- Completar traducción inglesa y firma de publicación Android.

Las pruebas de interfaz y datos de desarrollo usan harkai_test_* y servicios locales, sin insertar incidentes en producción.
