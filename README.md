# 💸 Expense Tracker

A secure, multi-user monthly expense-tracking web app for **Philippine Peso (PHP)** budgets. Each member sets their monthly income and logs expenses across four buckets — **Fixed, Variable, Wants, Debts** — and sees a live dashboard (total income, total expenses, savings target, actual savings, remaining budget) plus a pie chart of where the money goes. Admins set the global savings-target policy.

## Tech stack

| Layer | Technology |
| ----- | ---------- |
| Backend | Go (chi router) + **GORM** ORM |
| Database | MariaDB / MySQL |
| Frontend | React + Vite + TypeScript |
| Charts | Chart.js (`react-chartjs-2`) |
| Auth | JWT access tokens + rotating refresh tokens (HttpOnly cookie), bcrypt |

## Features

- 🔐 JWT auth with refresh-token rotation & server-side revocation, bcrypt hashing, per-IP rate limiting, `member`/`admin` roles
- 👤 Open self-registration; each user's data is ownership-scoped
- 🗂️ Categories typed as Fixed / Variable / Wants / Debts (a starter set is seeded per user)
- 🧾 Monthly expense CRUD with validation
- 📊 Dashboard KPIs + pie chart (by category or by bucket), per selected month
- ⚙️ Admin-configurable savings-target policy (percent of income or fixed ₱), applied to every member
- 🛡️ Admin user moderation — **suspend**, **ban**, **reactivate**, or **delete** members (blocking takes effect immediately and revokes active sessions; admins can't lock out their own account)

## Project structure

```
expense-tracker/
├── backend/          # Go JSON API
│   ├── cmd/server/           # entrypoint
│   └── internal/             # config, db, models, repository, auth, middleware, handlers, services
└── frontend/         # React + Vite SPA
    └── src/                  # api, auth, components, pages
```

## Prerequisites

- Go 1.24+
- Node 20+
- MariaDB / MySQL
- (Local, optional) [Laravel Valet](https://laravel.com/docs/valet) for the `https://expense-tracker.test` domain

## Environment files

Config is environment-specific. The backend loads `.env.<APP_ENV>` (falling back to a shared `.env`); real process env vars always win.

**Backend** (`backend/`)

| File | When | Committed? |
| ---- | ---- | ---------- |
| `.env.local` | `APP_ENV=local` (default) — local dev via Valet | ❌ gitignored |
| `.env.production` | `APP_ENV=production` — the VPS | ❌ gitignored |
| `.env.example` | reference template | ✅ tracked |

**Frontend** (`frontend/`) — Vite embeds `VITE_*` vars into the public bundle, so **no secrets** go here.

| File | When |
| ---- | ---- |
| `.env.development` | `npm run dev` |
| `.env.production` | `npm run build` |

## Local development (Laravel Valet)

Serve the app at **https://expense-tracker.test** by proxying the domain to the Vite dev server (which proxies `/api` to the Go backend):

```bash
# 1. Database
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS expense_tracker"   # prompts for your DB password

# 2. Backend (auto-migrates + seeds, serves :8080)
cd backend
cp .env.example .env.local        # then edit values (already provided in this repo)
APP_ENV=local go run ./cmd/server

# 3. Frontend (serves :5173, proxies /api → :8080)
cd frontend
npm install
npm run dev

# 4. Point Valet at the dev server and enable TLS (needs sudo)
valet proxy expense-tracker http://127.0.0.1:5173 --secure
```

Then open **https://expense-tracker.test**. To remove later: `valet unproxy expense-tracker`.

> The account registered with `ADMIN_EMAIL` (in `.env.local`) is auto-promoted to **admin**.

## Production deployment (VPS)

> **Full step-by-step guide for AlmaLinux 9 + Apache: [`DEPLOYMENT.md`](DEPLOYMENT.md).**
> Ready-made configs are in `deploy/` (systemd unit, Apache vhost) and `deploy.sh` (pull → build → restart). Quick outline below.

1. **Build the frontend** and serve the static files:
   ```bash
   cd frontend && npm ci && npm run build   # outputs dist/
   ```
2. **Build the backend** binary:
   ```bash
   cd backend && go build -o server ./cmd/server
   ```
3. **Configure** `backend/.env.production` — fill in every `CHANGE_ME` (strong `JWT_SECRET` via `openssl rand -base64 48`, a dedicated non-root DB user, real `CORS_ORIGINS`). Keep this file only on the VPS.
4. **Run** the API (behind a process manager like systemd):
   ```bash
   APP_ENV=production ./server
   ```
5. **Reverse proxy** (nginx/Caddy) on your domain over HTTPS:
   - serve the SPA `dist/` at `/`
   - proxy `/api` → `http://127.0.0.1:8080`

   Same-origin `/api` means the HttpOnly refresh cookie and `VITE_API_BASE_URL=` (empty) work unchanged. `COOKIE_SECURE=true` + HTTPS keep sessions secure.

Example nginx location block:

```nginx
root /var/www/expense-tracker/frontend/dist;
location / { try_files $uri /index.html; }
location /api/ { proxy_pass http://127.0.0.1:8080; proxy_set_header Host $host; }
```

## Testing

```bash
cd backend && go test ./...      # budget/savings logic (percent + fixed targets)
```

## API overview

| Method | Path | Auth | Purpose |
| ------ | ---- | ---- | ------- |
| POST | `/api/auth/register` | – | Create account (open self-registration) |
| POST | `/api/auth/login` | – | Log in |
| POST | `/api/auth/refresh` | cookie | Rotate refresh token, issue new access token |
| POST | `/api/auth/logout` | cookie | Revoke session |
| GET/PATCH | `/api/me` | ✔ | Profile / set monthly income |
| GET/POST | `/api/categories` | ✔ | List / create categories |
| PATCH/DELETE | `/api/categories/{id}` | ✔ | Edit / delete a category |
| GET/POST | `/api/expenses?month=YYYY-MM` | ✔ | List (by month) / create expenses |
| PATCH/DELETE | `/api/expenses/{id}` | ✔ | Edit / delete an expense |
| GET | `/api/dashboard?month=YYYY-MM` | ✔ | KPIs + category/bucket breakdown |
| GET | `/api/admin/users` | admin | List members |
| PATCH | `/api/admin/users/{id}/status` | admin | Suspend / ban / reactivate a member |
| DELETE | `/api/admin/users/{id}` | admin | Permanently delete a member + their data |
| GET/PUT | `/api/admin/settings` | admin | Read / set savings-target policy |

## Security notes

- **Passwords:** bcrypt (cost 12); hashes never leave the server. Registration enforces a policy — length 8–72, no common/breached passwords, and no email-name embedded in the password.
- **Tokens:** short-lived JWT access token (in memory on the client) + long-lived refresh token stored **hashed** server-side in an HttpOnly, `SameSite=Strict` cookie, rotated on every refresh and revocable on logout / suspend / ban. JWTs are verified against HMAC only (no algorithm-confusion). A weak `JWT_SECRET` (<32 chars) is rejected in production.
- **Authorization:** all member queries are ownership-scoped in the repository layer; admin routes gated by `RequireAdmin`; suspended/banned accounts are rejected on every request.
- **Rate limiting (layered, per client IP):** a broad global limiter on all `/api` routes, a tighter limiter on `/auth/*`, and a **per-email** brute-force throttle on login.
- **Anti-enumeration:** login returns an identical generic error and runs a constant-time bcrypt comparison whether or not the account exists.
- **Headers & transport:** every API response carries `Content-Security-Policy`, `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, and `Permissions-Policy`; the Apache vhost sets a strict SPA CSP + HSTS. `ENV=production` / `COOKIE_SECURE=true` enable `Secure` cookies and HSTS (serve behind TLS).
- **Hardening:** request body capped at 1 MB with unknown-field rejection; server has read/write/idle/header timeouts (Slowloris mitigation) and a header-size cap.
