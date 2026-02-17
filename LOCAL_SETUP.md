# LOCAL_SETUP.md - ProjectManagement

## 1. Prerequisites

- Go `1.21+`
- Node.js `18+` and `npm`
- Git

## 2. Clone and Enter Repo

```bash
git clone https://github.com/duncan-2126/ProjectManagement.git
cd ProjectManagement
```

## 3. Install/Sync Dependencies

```bash
go mod tidy
cd web && npm install && cd ..
```

## 4. Build and Verify

```bash
go test ./...
go build ./...
cd web && npm run build && cd ..
```

Expected result:
- Go tests pass
- Go build succeeds
- Web build succeeds (Vite may print a chunk-size warning; this is non-blocking)

## 5. First Successful Boot (CLI + Web)

### CLI

```bash
go run . init
go run . scan
go run . list
```

### Web GUI

```bash
go run . serve --host 127.0.0.1 --port 8080
```

Open:
- `http://127.0.0.1:8080`
- API health check by listing TODOs: `http://127.0.0.1:8080/api/todos`

## 6. Troubleshooting

- If `go test` fails with cache permission issues, run:

```bash
mkdir -p /tmp/gocache
GOCACHE=/tmp/gocache go test ./...
```

- If `todo serve` cannot find frontend assets, build the web app first:

```bash
cd web && npm run build && cd ..
```

- If your local document refers to `feature/CCDC`, note that this repository currently exposes feature branches such as `feature/web-gui`, `feature/due-dates`, and `feature/time-tracking`, but not `feature/CCDC`.
