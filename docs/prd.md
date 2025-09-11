📄 Draft PRD v1 — Open Mission Control

Product Name: Open Mission Control (IoT Telemetry & Alerting Platform)

Version: v1 (MVP, 10–12 weeks solo dev)

1. Goal

Build a multi-tenant SaaS platform for real-time IoT telemetry collection, visualization, and alerting. Think “Datadog for devices,” simplified but production-grade.

2. Target Users

IoT developers/teams that want to monitor fleets of devices.

Ops engineers who need real-time alerts and dashboards.

Small orgs/startups who can’t afford enterprise observability tools.

3. Core Features (MVP)

Multi-tenant organizations & teams (RBAC, invite, audit log).

Device telemetry ingestion via HTTP + MQTT.

Storage in time-series DB (TimescaleDB/ClickHouse).

Rule engine: simple thresholds → email/webhook alerts.

Next.js dashboards with live charts + rule builder UI.

Observability: traces, metrics, logs for self-dogfooding.

Security: OIDC login, TLS everywhere.

4. Nice-to-Haves (Stretch)

SMS alerts.

Billing/quotas (Stripe).

Multi-language/i18n.

Canary deployments with auto-rollback.

5. Non-Functional Requirements

Handle sustained 50k msgs/sec ingest.

p95 write latency <60ms, query latency <120ms.

99.9% availability (SLO target).

Tenant data isolation (no leaks across orgs).

6. Success Criteria

Demo with simulated devices sending data.

Live dashboards show streaming metrics.

Alerts fire correctly under load tests.

CI/CD pipeline with rollback works end-to-end.

7. Out of Scope for v1

Complex analytics (ML anomaly detection).

Native mobile apps (stick to PWA).

On-prem installations (cloud-only MVP).