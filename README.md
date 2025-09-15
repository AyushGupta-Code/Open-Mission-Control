# 🌌 Open Mission Control

This project sets up a **full-stack authentication & API gateway system** using:

- **Keycloak** (Identity Provider)
- **Next.js 15 + NextAuth** (Frontend)
- **Go API Gateway** (Backend)
- **Postgres** (Database)

So far, we have a working environment where:
- Users can log in with **Keycloak** from the Next.js frontend (`testuser` works ✅).
- The Go API gateway verifies tokens with Keycloak and exposes protected routes (`/me`, `/missions`).
- The backend can talk to Postgres (migrations will come later).

---

## ✅ What’s Done

- **Dockerized environment**  
  - Keycloak (`http://localhost:8081`)  
  - Postgres (`localhost:5432`)  
  - API Gateway in Go (`http://localhost:8080`)  
  - Next.js frontend (`http://localhost:3000`)  

- **Keycloak setup**
  - Realm: `open-mission-control`
  - Client: `open-mission-control-frontend`
  - Resource client: `omc-api`
  - User: `testuser` with password set

- **Frontend (Next.js + NextAuth)**
  - Login via Keycloak
  - Session handling with access & refresh tokens
  - API calls to backend with bearer token

- **Backend (Go API Gateway)**
  - Verifies JWT tokens from Keycloak
  - Endpoints:
    - `/healthz` → service alive
    - `/healthz/db` → Postgres check
    - `/me` → show decoded user claims
    - `/missions` (GET, POST, PUT, DELETE) → role-based mission management

---

## 📝 What’s Left

- [ ] Automate Keycloak setup (realm, clients, testuser) with `kcadm.sh` or import JSON config
- [ ] Add DB migrations for `missions` table
- [ ] Persist missions into Postgres (currently only SELECT works properly)
- [ ] Role-based UI in frontend (e.g. only admins can create/update/delete missions)
- [ ] Production-ready Docker setup (volumes, secrets, HTTPS)

---

## ▶️ How to Run

### 1. Start Infrastructure
Make sure you’re inside the project root:

```bash
docker-compose up -d keycloak postgres

Keycloak: http://localhost:8081

Postgres: postgres://omc:omc@localhost:5432/omc

2. Backend (Go API Gateway)

In another terminal:

cd backend/api-gateway
go run main.go


Runs on → http://localhost:8080

Test endpoints:

curl http://localhost:8080/healthz
curl -H "Authorization: Bearer <ACCESS_TOKEN>" http://localhost:8080/me

3. Frontend (Next.js)

In another terminal:

cd frontend
npm install
npm run dev


Runs on → http://localhost:3000

Click Sign in with Keycloak and log in with testuser.

4. Environment Variables

Create frontend/.env.local:

NEXTAUTH_SECRET=supersecretkey
NEXTAUTH_URL=http://localhost:3000

KEYCLOAK_ISSUER=http://localhost:8081/realms/open-mission-control
KEYCLOAK_CLIENT_ID=open-mission-control-frontend
KEYCLOAK_CLIENT_SECRET=<copy from Keycloak admin console>

🧭 Next Steps

Add DB migration scripts to automatically create missions table.

Move mock missions → Postgres persistence.

Role-based UI in frontend for mission CRUD.

Export Keycloak realm config to JSON so we can reimport it easily.

Write CI/CD pipeline to spin up the whole stack with one command.

🔑 Credentials (Dev only)

Keycloak Admin

URL: http://localhost:8081/admin

Username: admin

Password: admin

Test User

Username: testuser

Password: <set manually in Keycloak UI>
