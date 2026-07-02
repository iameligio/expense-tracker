# Deploying to a VPS — AlmaLinux 9 + Apache

This guide deploys the Expense Tracker on **AlmaLinux 9** with **Apache (httpd)** in front, the Go API kept alive by **systemd**, and updates via a manual **`git pull`**. Ready-made config files live in `deploy/` and `deploy.sh`.

**Architecture on the server**

```
Browser ──HTTPS──▶ Apache (:443) ──┬──▶ static SPA files  (frontend/dist)
                                   └──▶ /api  →  Go API (127.0.0.1:8080, systemd)
                                                      │
                                                      ▼
                                                 MariaDB (127.0.0.1:3306)
```

Because the SPA and `/api` share one domain, the HttpOnly refresh cookie stays same-origin — no CORS or cookie changes needed.

Run everything below as a sudo-capable user. Replace `expense-tracker.example.com` with your real domain everywhere.

---

## 1. Install prerequisites

```bash
sudo dnf update -y
sudo dnf install -y git httpd mariadb-server mod_ssl

# Go (toolchain — install the current release from go.dev to guarantee 1.24+)
GO_VER=1.24.5
curl -LO "https://go.dev/dl/go${GO_VER}.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf "go${GO_VER}.linux-amd64.tar.gz"
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
source /etc/profile.d/go.sh && go version

# Node.js 20 (for building the frontend)
curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash -
sudo dnf install -y nodejs
node --version
```

Enable the services:

```bash
sudo systemctl enable --now httpd mariadb
```

---

## 2. Open the firewall

```bash
sudo firewall-cmd --permanent --add-service=http --add-service=https
sudo firewall-cmd --reload
```

---

## 3. Configure the database

```bash
sudo mysql_secure_installation   # set a root password, answer the prompts

sudo mysql -u root -p <<'SQL'
CREATE DATABASE IF NOT EXISTS expense_tracker CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'expense_app'@'localhost' IDENTIFIED BY 'REPLACE_WITH_STRONG_DB_PASSWORD';
GRANT ALL PRIVILEGES ON expense_tracker.* TO 'expense_app'@'localhost';
FLUSH PRIVILEGES;
SQL
```

Use a **dedicated, non-root** DB user (`expense_app` above) — not the MariaDB root account.

---

## 4. Get the code

Create a service user and clone into `/var/www` (which already carries the correct SELinux context for Apache):

```bash
sudo useradd -r -m -d /var/www/expense-tracker -s /sbin/nologin expense || true
sudo mkdir -p /var/www/expense-tracker
sudo chown -R expense:expense /var/www/expense-tracker

# Clone as the service user (use your repo URL)
sudo -u expense git clone https://github.com/<you>/expense-tracker.git /var/www/expense-tracker
```

> If you push to a private repo, set up a deploy key or use HTTPS with a token so `sudo -u expense git pull` works non-interactively.

---

## 5. Configure production environment

Create `backend/.env.production` from the template and fill in every `CHANGE_ME`:

```bash
cd /var/www/expense-tracker/backend
sudo -u expense cp .env.example .env.production
sudo -u expense nano .env.production
```

Set at least:

```ini
APP_ENV=production
ENV=production
PORT=8080
CORS_ORIGINS=https://expense-tracker.example.com
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=expense_app
DB_PASSWORD=REPLACE_WITH_STRONG_DB_PASSWORD
DB_NAME=expense_tracker
JWT_SECRET=PASTE_OUTPUT_OF_openssl_rand_base64_48
ADMIN_EMAIL=etapiador@gmail.com
COOKIE_SECURE=true
```

Generate a strong secret:

```bash
openssl rand -base64 48
```

`.env.production` is gitignored, so it lives only on the server.

---

## 6. Build

```bash
cd /var/www/expense-tracker

# Backend binary (auto-migrates + seeds the DB on first run)
sudo -u expense bash -lc 'cd backend && go build -o server ./cmd/server'

# Frontend static bundle → frontend/dist
sudo -u expense bash -lc 'cd frontend && npm ci && npm run build'
```

The frontend is a **Progressive Web App**: `npm run build` also emits a service worker (`sw.js`, `workbox-*.js`), a `manifest.webmanifest`, and `registerSW.js` into `frontend/dist`. These are plain static files served straight from `DocumentRoot` by Apache — no extra build or config step. The app icons in `frontend/public/` are committed to the repo, so no icon-generation step runs on the server (`npm run generate-pwa-assets` is a local-only dev helper).

---

## 7. Run the API under systemd

```bash
sudo cp deploy/expense-tracker.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now expense-tracker

# Verify
sudo systemctl status expense-tracker --no-pager
curl -fsS http://127.0.0.1:8080/health && echo
```

If it fails to start, check `sudo journalctl -u expense-tracker -e` (usually a wrong DB password or missing `.env.production`).

---

## 8. Configure Apache

```bash
sudo cp deploy/apache-expense-tracker.conf /etc/httpd/conf.d/expense-tracker.conf
sudo sed -i 's/expense-tracker.example.com/YOUR_REAL_DOMAIN/' /etc/httpd/conf.d/expense-tracker.conf
```

**SELinux** — AlmaLinux blocks Apache from making network connections by default, so allow the proxy to reach the Go API:

```bash
sudo setsebool -P httpd_can_network_connect 1
```

Test the config and reload:

```bash
sudo apachectl configtest    # should print "Syntax OK"
sudo systemctl reload httpd
```

Your site should now answer on `http://YOUR_REAL_DOMAIN`.

> **PWA caching note.** This vhost sets no `Expires`/`Cache-Control` headers, so the service worker (`sw.js`) and `index.html` are revalidated normally and PWA updates propagate on the next visit — nothing to configure. If you later add far-future caching for the hashed, immutable bundles in `/assets/*`, **exclude `sw.js`, `registerSW.js`, `index.html`, and `manifest.webmanifest`** from it, or clients will get stuck on a stale app shell.

---

## 9. Enable HTTPS (Let's Encrypt)

```bash
sudo dnf install -y certbot python3-certbot-apache
sudo certbot --apache -d expense-tracker.example.com
```

certbot adds the `:443` vhost, installs the certificate, and sets up an HTTP→HTTPS redirect. Auto-renewal is handled by the `certbot-renew.timer`. Since `COOKIE_SECURE=true`, the refresh cookie now works correctly over TLS.

Open **https://expense-tracker.example.com**, register `etapiador@gmail.com` (auto-promoted to admin because it matches `ADMIN_EMAIL`), set your income, and start logging expenses.

---

## 10. Updating later (manual git pull)

A helper script rebuilds and restarts everything:

```bash
cd /var/www/expense-tracker
sudo -u expense git pull --ff-only          # or run the whole thing via deploy.sh
sudo -u expense ./deploy.sh
```

`deploy.sh` runs: `git pull` → rebuild Go binary → `npm ci && npm run build` → `systemctl restart expense-tracker` → `systemctl reload httpd` → health check. (The service user needs sudo rights for those two systemctl calls, or run just the restart/reload steps as your sudo user.)

Schema changes are applied automatically by GORM's `AutoMigrate` when the API restarts.

---

## 11. Automatic deploys with GitHub Actions (optional)

`.github/workflows/deploy.yml` SSHes into the VPS on every push to `main` and runs `deploy.sh` — the same script you use manually. Set it up once:

**a. Create a deploy key** (on your workstation or the VPS):

```bash
ssh-keygen -t ed25519 -f ~/.ssh/expense_deploy -N "" -C "github-actions-deploy"
```

Add the **public** key to the deploy user's `authorized_keys` on the VPS:

```bash
# run on the VPS, as the user GitHub will log in as
cat >> ~/.ssh/authorized_keys < /path/to/expense_deploy.pub
```

**b. Give that user permission to run the deploy.** The SSH user must:
- own `APP_PATH` (so `git pull` and the build write successfully), and
- be able to run the two service commands without a password prompt.

Add a sudoers rule (run `sudo visudo -f /etc/sudoers.d/expense-deploy`), replacing `deployuser`:

```
deployuser ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart expense-tracker, /usr/bin/systemctl reload httpd
```

> Simplest model: let this SSH user also **own the repo** and match the systemd `User=`. If you keep the service running as `expense`, either log in as `expense` (give it a shell + `authorized_keys`) or adjust `deploy.sh` to `sudo -u expense` the git/build steps.

**c. Add the repo secrets** — GitHub → repo **Settings → Secrets and variables → Actions → New repository secret**:

| Secret | Value |
| ------ | ----- |
| `SSH_HOST` | VPS IP or hostname |
| `SSH_USER` | the deploy user (e.g. `deployuser`) |
| `SSH_KEY` | contents of the **private** key `~/.ssh/expense_deploy` |
| `APP_PATH` | `/var/www/expense-tracker` |
| `SSH_PORT` | *(optional)* SSH port if not `22` |

**d. Push to `main`.** The Actions tab shows the run; on success the VPS has pulled, rebuilt, and restarted. You can still run `./deploy.sh` by hand anytime, and trigger a deploy manually from **Actions → Deploy to VPS → Run workflow**.

> First-connection note: `appleboy/ssh-action` accepts the host key automatically, so no `known_hosts` setup is needed.

---

## Troubleshooting

| Symptom | Likely cause / fix |
| ------- | ------------------ |
| 503 from Apache on `/api` | Go service down (`systemctl status expense-tracker`) or SELinux — run `setsebool -P httpd_can_network_connect 1` |
| API won't start | Wrong DB creds or missing `.env.production` — `journalctl -u expense-tracker -e` |
| SPA loads but refresh 404s | `FallbackResource /index.html` missing from the vhost `<Directory>` block |
| Login works but you're logged out on reload | Cookie not `Secure`/not sent — confirm HTTPS is live and `COOKIE_SECURE=true` |
| `git pull` prompts for credentials | Configure a deploy key or token for the `expense` user |
| App shows an old version after a deploy | Service worker serves the cached shell until it updates — it refreshes on the next visit/reload (`registerType: autoUpdate`). Force it via the browser's **Application → Service Workers → Update/Unregister**. Don't add long-lived cache headers to `sw.js`/`index.html` (see the PWA caching note in §8). |
| "Install app" option missing in the browser | PWA install needs HTTPS (finish §9) — it won't appear over plain `http://`. |
