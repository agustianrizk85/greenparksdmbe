# SDM & Legal API — Legal-Permit War Room

Go (stdlib-only) backend for the Greenpark **SDM & Legal** "Legal-Permit War
Room". It exposes generic CRUD over the legal/permit data collections (projects,
units, BPN, PBG, PKS bank, documents, risks, bottlenecks, escalations, evidence,
actions, daily, weekly) plus token-based auth.

Data is seeded from an embedded snapshot (`internal/repository/seed.json`,
exported from the design) and persisted to a JSON file so edits survive restarts.
Records are heterogeneous JSON objects keyed by `id`.

## Architecture

```
cmd/server            composition root
internal/
  config              env-based configuration (SDM_PORT, ...)
  passwd / auth       salted SHA-256 hashing + in-memory bearer-token sessions
  domain              User + generic Record + collection allow-list
  repository          embedded seed.json + file-persisted generic store
  service             CRUD + auth use-cases
  transport/http      router, handlers, middleware, JSON helpers
```

## Auth & roles

- `POST /api/auth/login` → `{ token, user }`. Send `Authorization: Bearer <token>`.
- **admin / admin123** — full CRUD. **viewer / viewer123** — read-only (writes → 403).

## API

| Method · Path                 | Description                              |
| ----------------------------- | ---------------------------------------- |
| `GET /api/health`             | Liveness probe (public)                  |
| `POST /api/auth/login`        | Login (public)                           |
| `GET /api/auth/me`            | Current user (auth)                      |
| `POST /api/auth/logout`       | Revoke token (auth)                      |
| `GET /api/data`               | All collections in one map (auth)        |
| `GET /api/{col}`              | List a collection (auth)                 |
| `POST /api/{col}`             | Create a record (admin)                  |
| `PATCH /api/{col}/{id}`       | Merge-update a record (admin)            |
| `PUT /api/{col}/{id}`         | Update a record (admin)                  |
| `DELETE /api/{col}/{id}`      | Delete a record (admin)                  |

Collections: `projects, units, bpn, pbg, pks, documents, risks, bottlenecks,
escalations, evidence, actions, daily, weekly`.

## Run

```bash
cd backend/sdm
go run ./cmd/server
# sdm API listening on http://localhost:8087
```

| Variable           | Default                  | Description                  |
| ------------------ | ------------------------ | ---------------------------- |
| `SDM_PORT`         | `8087`                   | HTTP port                    |
| `SDM_ALLOW_ORIGIN` | `*`                      | CORS allowed origin          |
| `SDM_DATA_PATH`    | `data/sdm-data.json`     | JSON file the data persists to |
