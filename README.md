# Coding Challenge — Go + Fiber y Node.js + Express

Implementación inicial del reto técnico del PDF.

## Arquitectura

- **Go + Fiber** (`go-api`): recibe una matriz, calcula su factorización QR y consulta por HTTP a la API Node.js.
- **Node.js + Express** (`node-api`): recibe las matrices Q y R y calcula máximo, mínimo, promedio, suma total y si cada matriz es diagonal.
- **Docker Compose**: construye y ejecuta ambas API en una red común.

## Nota sobre el enunciado

El documento menciona tanto “rotación de la matriz” como “factorización QR”. Son operaciones diferentes. Esta implementación prioriza la factorización QR porque figura explícitamente en la sección “Funcionalidad requerida”. La operación adicional calcula estadísticas sobre Q y R. Esta decisión debe explicarse al entrevistador y, si es posible, confirmarse con RR. HH.

## Requisitos para ejecutar

- Windows 10/11 compatible con Docker Desktop y WSL 2 habilitado.
- Docker Desktop instalado y en ejecución.
- Conexión a Internet durante la primera construcción para descargar imágenes y dependencias.

No hace falta instalar Go ni Node.js directamente en Windows para ejecutar con Docker.

## Iniciar el proyecto

Desde PowerShell o la terminal integrada de VS Code, en la carpeta raíz del proyecto:

```powershell
docker compose up --build
```

La primera ejecución puede tardar porque descarga imágenes y dependencias. Para detener los servicios: `Ctrl+C`. Para eliminar los contenedores: `docker compose down`.

## Probar la API

### 1. Verificar estado

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:3000/health
```

### 2. Solicitar factorización QR y estadísticas

En PowerShell:

```powershell
$body = @{
  matrix = @(
    @(1, 2)
    @(3, 4)
    @(5, 6)
  )
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri http://localhost:8080/factorize `
  -Method Post -ContentType "application/json" -Body $body
```

La respuesta incluye `q`, `r` y `statistics`.

También se puede usar `curl`:

```bash
curl -X POST http://localhost:8080/factorize \
  -H "Content-Type: application/json" \
  -d "{\"matrix\":[[1,2],[3,4],[5,6]]}"
```

## Endpoints

### Go API — `http://localhost:8080`

- `GET /health`: comprobación de estado.
- `POST /factorize`: recibe `{ "matrix": [[...], [...]] }`, calcula QR y consulta a Node.

### Node API — `http://localhost:3000`

- `GET /health`: comprobación de estado.
- `POST /stats`: recibe `{ "matrices": { "Q": [[...]], "R": [[...]] } }` y devuelve estadísticas por matriz y globales.

## Decisiones y límites

- Se exige una matriz rectangular no vacía con filas del mismo tamaño.
- Para la factorización QR se requiere número de filas mayor o igual al número de columnas (`m >= n`).
- Los números deben ser finitos.
- La factorización usa Gonum. Debido a la naturaleza numérica, pueden aparecer pequeñas diferencias de redondeo.
- “Matriz diagonal” se evalúa en cada matriz por separado; para matrices no cuadradas se devuelve `false`.
- Las estadísticas globales consideran todos los valores de Q y R.
- No se han incluido las funcionalidades opcionales de frontend, JWT ni pruebas automatizadas.

## Verificación matemática

Para la salida QR, Q tiene columnas ortonormales y R es triangular superior. Se cumple aproximadamente `A = Q × R`, con pequeñas diferencias por redondeo.
