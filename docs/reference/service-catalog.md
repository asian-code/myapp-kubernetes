# Service Catalog

This catalog lists all microservices currently running in the myHealth platform.

---

## 🟢 api-service

**Description**: The main entry point for the platform. Handles HTTP REST requests from the frontend/mobile app.
*   **Language**: Go 1.24
*   **Type**: REST API
*   **Port**: 8080
*   **Dependencies**: PostgreSQL (RDS)
*   **Code Location**: [`services/api-service`](../../services/api-service)
*   **Owner**: Platform Team

### Key Endpoints
*   `GET /health`: Health check
*   `POST /auth/login`: User authentication
*   `GET /users/{id}`: Get user profile

---

## 🔵 data-processor

**Description**: Asynchronous background worker. Consumes messages from the queue and processes health data (e.g., calculating daily scores).
*   **Language**: Go 1.24
*   **Type**: Worker (Non-HTTP)
*   **Dependencies**: PostgreSQL (RDS), SQS (or internal queue)
*   **Code Location**: [`services/data-processor`](../../services/data-processor)
*   **Owner**: Data Team

---

## 🟣 oura-collector

**Description**: Scheduled job that fetches data from the Oura Ring API.
*   **Language**: Go 1.24
*   **Type**: CronJob
*   **Schedule**: Every 6 hours
*   **Dependencies**: Oura External API, PostgreSQL
*   **Code Location**: [`services/oura-collector`](../../services/oura-collector)
*   **Owner**: Integrations Team
