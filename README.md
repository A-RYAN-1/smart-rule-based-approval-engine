# Smart Rule-Based Approval Engine

[![CI](https://github.com/A-RYAN-1/smart-rule-based-approval-engine/actions/workflows/ci.yml/badge.svg)](https://github.com/A-RYAN-1/smart-rule-based-approval-engine/actions/workflows/ci.yml)
[![Uptime](https://github.com/A-RYAN-1/smart-rule-based-approval-engine/actions/workflows/uptime.yml/badge.svg)](https://github.com/A-RYAN-1/smart-rule-based-approval-engine/actions/workflows/uptime.yml)

Employees request leave, expense reimbursements and discounts; a rules engine decides each request the moment it arrives. Requests inside the employee's grade rules are approved automatically, the rest go to their manager, and a nightly job auto-rejects requests left undecided for 7 working days. Admins manage the rules, holidays and reports.

**Live:** https://smart-rule-based-approval-engine.vercel.app

---

## Try It Live

Open the [live app](https://smart-rule-based-approval-engine.vercel.app) and sign in with one of the demo accounts (the login page lists them too and fills the form for you):

| Role | Email | Password | Try |
|---|---|---|---|
| Employee | `employee@company.com` | `employee-demo` | Request 2 days of leave (approved instantly by the rules) and 10 days (waits for the manager) |
| Manager | `manager@company.com` | `manager-demo` | Approve or reject the employee's pending requests |
| Admin | `admin@company.com` | `admin-demo` | Change a rule's limit, add a holiday, see the reports |

These are public demo accounts with sample data. Each night the rules go back to the defaults, so whatever a visitor changes with the admin account is undone. You can also sign up as a new employee.

---

## Features

- **Three request types**: leave (counted in calendar days, from and to inclusive), expense (amount and category) and discount (percentage), each with a balance per employee.
- **Rules engine**: one rule per request type and grade, as JSON conditions such as `{"max_days": 3}`, `{"max_amount": 5000}` or `{"max_percent": 10}`. Inside the limit: `AUTO_APPROVED`; above it: `PENDING` for the manager.
- **Approvals**: managers and admins see pending requests, approve or reject with a comment; employees can cancel their own pending requests.
- **Nightly auto-reject** of requests still pending after 7 working days (weekends and admin-set holidays skipped), at midnight India time.
- **Roles**: employee, manager, admin, enforced on the server from the signed login token.
- **Admin**: create, edit and disable rules; manage holidays; status and request-type reports; user overview page.

Default rules (from the migrations):

| Grade | Leave auto-approved up to | Expense up to | Discount up to |
|---|---|---|---|
| Grade 1 (new employees) | 3 days | 5,000 | 10% |
| Grade 2 (manager) | 5 days | 10,000 | 20% |
| Grade 3 (admin) | 10 days | 20,000 | 30% |

---

## Architecture

```
Browser ── React SPA (Vite, TypeScript, TanStack Query)
   │  same origin: /api/*
   ▼
Go API (Gin) ── handlers → services (rules engine) → repositories (pgx)
   │                                                     │
   │  Vercel Function: api/index.go                      ▼
   │  Local server:    backend/cmd/server        PostgreSQL (Neon on Vercel,
   │                                             Docker locally)
   └─ Vercel Cron, daily ──► /api/cron/auto-reject
```

The backend is layered: HTTP handlers validate input, services hold the business rules (balance checks, rule evaluation, approvals in a transaction), repositories run SQL through `pgx`. `server.New` builds the whole app from a `Config`, so the same code runs as a long-lived server locally and as a Vercel Function in production. On startup it applies the SQL migrations embedded in the binary (`backend/migrations`, each in a transaction) and creates the admin, manager and demo employee accounts if they are missing. With `DEMO_MODE=true` the login page lists those accounts (`/api/demo-logins`) and the daily cron run restores the default rules. Sign-up only ever creates employees.

Configuration is read from environment variables in one place (`backend/config/config.go`), which refuses to start with a clear message when something required is missing or unsafe (a short `JWT_SECRET`, `CORS_ORIGINS=*`). Login sets an httpOnly, SameSite=Lax, Secure cookie holding a signed JWT. Every request gets a request id, used in one log line per request; server errors are logged with it and the client only sees a generic message.

---

## Run It Locally

**Prerequisites:** Git, [mise](https://mise.jdx.dev) (installs the pinned Go and Node versions; or install Go 1.26 and Node 22 yourself) and Docker (for PostgreSQL).

1. **Clone and install the toolchain**
   ```bash
   git clone https://github.com/A-RYAN-1/smart-rule-based-approval-engine.git
   cd smart-rule-based-approval-engine
   mise install          # reads mise.toml: Go 1.26, Node 22
   ```

2. **Start PostgreSQL**
   ```bash
   docker run -d --name approval-pg -p 5433:5432 \
     -e POSTGRES_USER=approval -e POSTGRES_PASSWORD=approval -e POSTGRES_DB=approval_engine \
     postgres:17-alpine
   ```

3. **Configure and start the backend** (terminal 1)
   ```bash
   cd backend
   cp .env.example .env
   # edit .env: set JWT_SECRET (openssl rand -hex 32) and the two SEED_*_PASSWORD values
   mise exec -- go run ./cmd/server
   ```
   It applies the migrations, creates `admin@company.com` and `manager@company.com` with your seed passwords, and listens on http://localhost:8080 (`curl localhost:8080/health`).

4. **Start the frontend** (terminal 2)
   ```bash
   cd frontend
   mise exec -- npm ci
   mise exec -- npm run dev
   ```
   Open http://localhost:5173. The dev server proxies `/api` to the backend, so no frontend configuration is needed. Sign up as an employee, or log in as the manager or admin.

---

## Environment Variables

Backend (`backend/.env.example` lists each with an explanation):

| Variable | Required | What it is |
|---|---|---|
| `DATABASE_URL` | yes | PostgreSQL connection string. Separate `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` are also accepted. |
| `JWT_SECRET` | yes | Signs login tokens; at least 32 characters, secret. |
| `SEED_ADMIN_PASSWORD`, `SEED_MANAGER_PASSWORD` | for approvals | Passwords for the built-in admin and manager, created on first start. |
| `SEED_EMPLOYEE_PASSWORD` | no | Password for a built-in `employee@company.com`, created on first start. |
| `DEMO_MODE` | no (default `false`) | `true` publishes the seed logins on the login page and resets the rules on each cron run. Only for a public demo: the seed passwords become public. |
| `COOKIE_SECURE` | no (default `true`) | `false` only for local http. |
| `CORS_ORIGINS` | no | Extra browser origins allowed to call the API, comma-separated. Empty means same-origin only. |
| `CRON_SECRET` | on Vercel | Protects `/api/cron/auto-reject`; Vercel Cron sends it automatically. |
| `GIN_MODE` | no (default `release`) | `release` or `debug`. |
| `PORT` | no (default `8080`) | Port of the local server. |
| `RUN_CRON` | no (default `true`) | Run the nightly job inside the local server. |

Frontend (`frontend/.env.example`): `VITE_API_URL`, empty by default (the app calls `/api` on its own origin).

---

## Tests and Checks

| Command (from) | What it does |
|---|---|
| `go test ./...` (backend) | Unit tests (services with mocks, config, JWT, rule helpers) |
| `TEST_DATABASE_URL=postgres://... go test ./server/` (backend) | Integration test of the main flow on a real, empty database: migrations, seed, sign-up, auto-approval, manager approval, role checks, cron auth, demo logins and the demo rule reset |
| `go vet ./...`, `gofmt -l .` (backend) | Static checks |
| `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` (backend) | Known vulnerabilities in reachable code |
| `npm run lint`, `npm test`, `npm run build` (frontend) | ESLint, Vitest, production build |

GitHub Actions runs all of them on every push, with PostgreSQL as a service container (`.github/workflows/ci.yml`), and checks the live site and its database once a day (`uptime.yml`).

Dependency status: `govulncheck` reports no vulnerabilities. `npm audit --omit=dev` reports one moderate React Router advisory (open redirect through `<Link>`/`navigate` targets); every navigation target in this app is a fixed path, so it cannot be triggered from user input. The other `npm audit` findings are in build-time tools (Vite, Tailwind, Vitest) that do not ship to the browser.

---

## Deployment (free)

Everything runs on Vercel's free Hobby plan as one project: the Vite build is served as static files, `api/index.go` is a Go Function serving every `/api/*` route, Vercel Cron calls the auto-reject job daily at 18:30 UTC (midnight India time), and the database is Neon's free plan through the Vercel Marketplace. `vercel.json` holds the build, routes, cron and security headers.

To host your own copy:

1. Import the repository in Vercel (root directory: the repository root; `vercel.json` sets the rest).
2. Add the Neon integration (Storage tab); it sets `DATABASE_URL`.
3. Set `JWT_SECRET`, `CRON_SECRET`, `SEED_ADMIN_PASSWORD` and `SEED_MANAGER_PASSWORD` in Settings, Environment Variables (plus `SEED_EMPLOYEE_PASSWORD` and `DEMO_MODE=true` for a public demo like this one).
4. Deploy. The first request applies the migrations and creates the seed accounts.

The same backend also runs anywhere a Go binary can (`go build ./cmd/server`): it reads `PORT` and runs the nightly job in-process.

---

## Project Structure

```
api/index.go                 Vercel Function entrypoint (wraps the Gin router)
vercel.json                  build, routes, cron, security headers
backend/
  cmd/server/                local server: graceful shutdown, in-process cron, `migrate up|create`
  server/                    builds the app: DB, migrations, seed, services, routes, middleware
  config/                    every environment variable, validated
  app/                       handlers and services per feature (auth, leave, expense,
                             discount, rules, holidays, reports, auto-reject, my requests)
  repositories/              SQL through pgx; repositories/migrations runs the migrations
  migrations/                SQL migrations, embedded into the binary
  seed/                      creates the seed accounts; demo logins and the demo rule reset
  pkg/                       middleware (JWT, roles), errors, responses, utils
frontend/src/
  pages/                     dashboard, request forms, my requests, approvals, admin pages
  services/, lib/api.ts      typed API client
  hooks/, contexts/          TanStack Query hooks, auth context
```

API reference: [`backend/api/swagger.yaml`](backend/api/swagger.yaml) and the Postman collection [`backend/API.txt`](backend/API.txt).

---

## Contributors

Built by [Ankita Advitot](https://github.com/Ankita-Advitot) and [Aryan Deshmukh](https://github.com/A-RYAN-1).

- **Ankita Advitot**: most of the Go backend: the rules engine, request and approval services, repositories, auto-reject job and the API.
- **Aryan Deshmukh**: the React frontend, deployment, and the production pass (secure configuration, migrations that work on a fresh install, tests and CI, the single Vercel deploy, demo accounts).

---

## License

[MIT](LICENSE)
