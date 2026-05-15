# Chat App

## Run

1. Start PostgreSQL:
   ```bash
   docker compose up -d
   ```

2. Export env:
   ```bash
   cp .env.example .env
   ```

3. Apply migrations with `migrate`:
   ```bash
   task migrate-up
   ```

4. Install Go deps and run:
   ```bash
   go mod tidy
   task run
   ```

Open `http://localhost:8080`.

## Notes

- Auth uses secure HTTP-only cookies.
- Chat updates are loaded with polling.
- Replace `COOKIE_SECURE=false` with `true` behind HTTPS.
