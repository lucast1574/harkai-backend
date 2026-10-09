# Harkai: alcance acordado

Esta especificación reúne las instrucciones del propietario. Es trabajo en curso; no describe capacidades ya desplegadas.

## Arquitectura

- Copias locales en `/home/lucast/Documents/code/harkai/`: app, front y backend.
- Backend Go común, modular y funcional, con funciones que retornan resultados, archivos de producción menores de 200 líneas, MongoDB y MinIO privado en HDD (cuota actual: 10 GiB).
- Firebase exclusivamente para Google login y entrega push con FCM. Identidad local, roles, datos, selección de alertas y verificación permanecen en Go. No Firestore ni Gemini.
- Clasificación de texto y ayuda por reglas; audio con reconocimiento local y confirmación editable. Las sugerencias nunca verifican un hecho.
- Moderación local de imágenes con OpenCV y detector entrenado. Rechazar antes de publicar/almacenar si la revisión falla; solo evidencia aprobada puede aparecer en el mapa. No prometer detección perfecta.
- Mapas Flutter/Leaflet con proveedor compatible y atribución. Eliminar Google Maps y su dependencia de claves.

## Reportes y notificaciones

- Retirar eventos. Mantener mascotas y lugares de ayuda; incorporar extorsión, secuestro, violencia y homicidio junto con categorías de emergencia existentes.
- Publicar inicialmente como **No verificado** y notificar también estos reportes según zona, radio y preferencias. No condicionar la entrega a una aprobación.
- Una aprobación de otra cuenta basta para **Confirmado por la comunidad**. El autor no puede confirmar su reporte. Escritura atómica: solo un aprobador, sin votos acumulados ni apariencia de verificación oficial.
- FCM es transporte, no base de datos. Go debe autenticar dispositivos, seleccionar destinatarios por zona y preferencias y mantener entrega/reintentos duraderos. Los payloads excluyen contactos privados y describen el estado no verificado.
- Al abrir la notificación, cargar el reporte vigente desde Go. Revocar dispositivos al cerrar sesión; no asociar un token a otra identidad sin autenticación. Móvil Android y web deben registrar permisos y tokens del proyecto nuevo.

## Roles y paneles web

- Roles: `user`, `gov`, `admin`. Registro siempre `user`; proteger cada operación en el backend, no solo ocultar componentes.
- Usuario: mapa, gráficos, reportes, noticias desbloqueadas y detalle con estado claro.
- Gobierno: vistas compartidas, reportes, mapas de calor y estadísticas útiles para municipios; filtros y exportaciones descargables. Rotular periodos, cobertura, reportes no verificados y datos parciales.
- Admin: funciones de gobierno, usuarios y asignación de roles; ganancias sin pagos activos ni cifras inventadas. Remover GPay y bloqueos por suscripción; Stripe queda para una fase posterior.
- Proteger administradores existentes contra pérdida accidental/concurrente de acceso; no permitir cambios de rol propios. Registrar quién asignó un rol y cuándo.
- Exportaciones no incluyen contactos privados; limitar consultas, escapar celdas y hacer visible cualquier paginación/límite.

## UX obligatoria

- **Sin barra superior de navegación con opciones.** Web: sidebar; móvil: menú lateral deslizable. En pantallas estrechas usar botón lateral para abrirlo.
- Componentes reutilizables, inputs custom accesibles con etiquetas, errores, estados de carga, teclado y enfoque correctos.
- Rediseñar el formulario/modal de reporte por pasos: texto/audio, categoría sugerida editable, ubicación, evidencia y revisión.
- Conservar funcionalidades útiles de perfil, preferencias, contactos, ayuda, mascotas, lugares, mapa, historial y compartir; reemplazar la infraestructura antes de retirar flujos antiguos.

## Pendiente de entrega

Conectar Google y FCM del proyecto nuevo; moderación/STT; reestructurar clientes y sidebar; paneles y exportación; pruebas de permisos, concurrencia, recepción push, builds de clientes y despliegue. No anunciar estos pasos como completos sin comprobarlos.

## Landing independiente

Repositorio `harkai-landing`, copia local junto a los otros repos y proyecto propio en Dokploy. `harkai.lat` es la landing; `panel.harkai.lat` aloja la aplicación web. Mantener branding existente y estilo moderno inspirado en Nexode, con navegación lateral. No inventar métricas, pagos, usuarios ni descargas disponibles.
