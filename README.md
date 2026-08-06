# geoquiz-api

Go (Gin) API for Geoquiz — countries, auth, scores, and admin invite code.

## Run (local)

Requires PostgreSQL. Copy `.env.example` and set `DATABASE_URL`, `JWT_SECRET`, `INVITE_CODE`, `ADMIN_EMAIL`.

```bash
go run ./cmd/server
```

## Docker

From this repo (expects sibling `../geoquiz` for the web image):

```bash
docker compose up --build
```

Open http://localhost:8080. Set `JWT_SECRET`, `INVITE_CODE`, and `ADMIN_EMAIL` via environment or a `.env` file next to `docker-compose.yml`.

## Main endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health` | Liveness |
| GET | `/api/v1/countries` | Country list |
| GET | `/api/v1/countries/geojson` | GeoJSON |
| POST | `/api/v1/auth/register` | Sign up (invite code) |
| POST | `/api/v1/auth/login` | Login |
| GET | `/api/v1/auth/me` | Current user |
| PATCH | `/api/v1/account` | Update profile |
| GET | `/api/v1/profiles/:username` | Public profile |
| GET | `/api/v1/scores` | Score-board (`?quiz_type=flag|map`, optional `limit`) |
| POST | `/api/v1/scores` | Save quiz score (auth) |
| GET/PUT | `/api/v1/admin/invite-code` | Admin invite code |
