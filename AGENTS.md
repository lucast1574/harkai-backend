# Harkai backend

- Go funcional: reglas puras en internal/core, orquestación en internal/application, efectos en adaptadores.
- Cada función de producción devuelve un resultado o error. Únicas excepciones: main y ServeHTTP, exigidas por Go. Los callbacks de pruebas también respetan la firma de testing.
- Mantén cada archivo Go de producción por debajo de 200 líneas. Separa responsabilidades, no fragmentes arbitrariamente.
- No añadas HTTP, MongoDB, entorno o proveedores externos al núcleo. Pasa tiempo e identidad explícitamente.
- MongoDB v2 oficial; concurrencia mediante escrituras condicionales e índices, nunca mutex locales como garantía entre réplicas.
- Firebase se limita al inicio de sesión con Google y a entregar notificaciones FCM del proyecto nuevo. Datos, permisos, reportes, verificación, selección de destinatarios y moderación pertenecen al backend Go; no usar Firestore ni Gemini. Ambas excepciones están autorizadas por el usuario.
- Las credenciales vienen del entorno; nunca imprimirlas ni incluirlas en repositorio, pruebas o documentación.
- API común versionada para web y móvil, contratos explícitos y errores estables.
- Antes de entregar: gofmt, go test -race ./..., go vet ./..., go build ./cmd/api y go run ./cmd/checkstructure.
- Integración con TEST_MONGODB_URI únicamente en bases harkai_test_* temporales; no usar la base de producción.
