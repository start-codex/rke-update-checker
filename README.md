# RKE Update Checker

Una herramienta de línea de comandos para verificar actualizaciones disponibles de charts Helm en clusters Kubernetes administrados por Rancher.

## Características

- ✅ Lista todas las aplicaciones Helm instaladas en todos los clusters
- 🔍 Compara versiones actuales con las últimas disponibles
- 📊 Muestra resultados en formato tabla organizado
- 🔧 Identifica charts administrados internamente por Rancher
- ⚡ Optimizado para evitar búsquedas duplicadas
- 🌐 Soporta múltiples repositorios de charts

## Requisitos

- Go 1.19 o superior
- Acceso a una instancia de Rancher
- Token de autenticación de Rancher con permisos de lectura

## Instalación

```bash
git clone <repository-url>
cd go-rancher
go build -o rke-update-checker cmd/rke-update-checker/main.go
```

## Configuración

La aplicación requiere las siguientes variables de entorno:

```bash
export RANCHER_URL="https://your-rancher-instance.com/v3"
export RANCHER_TOKEN="token-xxxxxxxxxxxxxxxx:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export VERBOSE="true"  # Opcional: muestra información detallada de procesamiento
```

### Obtener el Token de Rancher

1. Accede a tu instancia de Rancher
2. Ve a **User Settings** (configuración de usuario)
3. En la pestaña **API Keys**, crea un nuevo token
4. Copia el token generado

## Uso

### Ejecución directa

```bash
export RANCHER_URL="https://your-rancher-instance.com/v3"
export RANCHER_TOKEN="your-token-here"
go run cmd/rke-update-checker/main.go
```

### Ejecución con binario compilado

```bash
# Compilar
go build -o rke-update-checker cmd/rke-update-checker/main.go

# Configurar variables de entorno
export RANCHER_URL="https://your-rancher-instance.com/v3"
export RANCHER_TOKEN="your-token-here"
export VERBOSE="true"

# Ejecutar
./rke-update-checker
```

## Salida

La aplicación muestra una tabla con la siguiente información:

| Columna | Descripción |
|---------|-------------|
| CLUSTER | Nombre del cluster Kubernetes |
| NAMESPACE | Namespace donde está instalada la aplicación |
| RELEASE | Nombre del release de Helm |
| REPO | Repositorio del chart |
| CHART | Nombre del chart |
| CURRENT | Versión actualmente instalada |
| LATEST | Última versión disponible |
| STATUS | Estado del deployment |
| UPDATE | Estado de actualización disponible |
| SOURCES | URLs de origen del chart |

### Estados de Actualización

- ✅ **UP-TO-DATE**: La versión instalada es la más reciente
- ⚠️ **UPDATE AVAILABLE**: Hay una nueva versión disponible
- 🔧 **MANAGED**: Chart administrado internamente por Rancher
- ❓ **NOT FOUND**: No se pudo determinar la versión más reciente

## Arquitectura

El proyecto sigue las convenciones estándar de Go con la siguiente estructura:

```
go-rancher/
├── cmd/
│   └── rke-update-checker/
│       └── main.go              # Punto de entrada de la aplicación
├── internal/
│   ├── chart/
│   │   ├── chart.go             # Estructuras y lógica de charts
│   │   └── fetcher.go           # Obtención de charts desde repositorios
│   ├── display/
│   │   └── display.go           # Formateo y presentación de resultados
│   ├── helm/
│   │   └── release.go           # Decodificación de releases de Helm
│   ├── rancher/
│   │   ├── client.go            # Cliente principal de Rancher
│   │   └── internal_charts.go   # Manejo de charts internos
│   └── version/
│       └── version.go           # Comparación semántica de versiones
├── go.mod
├── go.sum
└── README.md
```

### Paquetes

- **rancher**: Cliente principal y orchestración de clusters
- **helm**: Decodificación y procesamiento de releases de Helm
- **chart**: Fetching y comparación de charts desde repositorios
- **version**: Comparación semántica de versiones
- **display**: Formateo y presentación de resultados

## Charts Administrados Internamente

Los siguientes charts son administrados automáticamente por Rancher y se marcan como "MANAGED":

- `fleet-agent-local`
- `rancher-provisioning-capi`
- `rancher-webhook`
- `system-upgrade-controller`
- `rke2-canal`
- `rke2-coredns`
- `rke2-ingress-nginx`
- `rke2-metrics-server`
- `rke2-runtimeclasses`
- `rke2-snapshot-controller`
- `rke2-snapshot-controller-crd`

## Troubleshooting

### Error: "RANCHER_URL environment variable is required"

Asegúrate de haber configurado la variable de entorno `RANCHER_URL` con la URL completa de tu instancia de Rancher incluyendo `/v3`.

### Error: "RANCHER_TOKEN environment variable is required"

Configura la variable de entorno `RANCHER_TOKEN` con un token válido de Rancher.

### Error de conexión SSL

Si tu instancia de Rancher usa certificados auto-firmados, la aplicación está configurada para saltarse la verificación SSL automáticamente.

### Sin resultados

Si no aparecen resultados:

1. Verifica que tienes acceso a los clusters
2. Asegúrate de que hay aplicaciones Helm instaladas
3. Ejecuta con `VERBOSE=true` para ver más detalles del procesamiento

## Contribuir

1. Fork el repositorio
2. Crea una rama para tu feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit tus cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Crea un Pull Request

## Licencia

[Agregar información de licencia aquí]