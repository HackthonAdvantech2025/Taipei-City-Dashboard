# Local Deployment Design Specification

## 1. Overview
The goal is to deploy the Taipei-City-Dashboard project locally using Docker Compose. This approach leverages the existing Docker configurations to create a development-friendly environment that closely mirrors the production setup.

## 2. Architecture
The deployment will be split into two main Compose files to separate stateful data services from stateless application services.

*   **Network:** A custom Docker bridge network (`br_dashboard`) will be created to allow communication between all containers.
*   **Data Layer (`docker-compose-db.yaml`):**
    *   PostgreSQL (PostGIS) for Dashboard data (`postgres-data`).
    *   PostgreSQL (PostGIS) for Manager data (`postgres-manager`).
    *   Redis for caching.
    *   Qdrant for vector search (AI features).
    *   pgAdmin for database management.
*   **Application Layer (`docker-compose.yaml`):**
    *   Go Backend (`dashboard-be`): Running in development mode with live reloading (if supported by the command) or standard run.
    *   Vue Frontend (`dashboard-fe`): Running via Vite dev server.
    *   Nginx: Serving as a reverse proxy (if applicable, though frontend dev server usually handles its own port).

## 3. Configuration
*   **Environment Variables:** A `.env` file will be created in the `docker/` directory based on `.env.template`. Essential API keys (Mapbox, TWCC) will need to be provided by the user for full functionality, but the system will be configured to start even without them (graceful degradation).
*   **Volumes:** Source code directories (`Taipei-City-Dashboard-BE` and `Taipei-City-Dashboard-FE`) will be mounted as volumes into their respective containers to enable local development without rebuilding images.

## 4. Execution Steps
1.  Create the Docker network: `docker network create br_dashboard` (with specific subnet if required by the template).
2.  Copy `.env.template` to `.env` and fill in default/dummy values for required fields to ensure startup.
3.  Start the data layer: `docker compose -f docker/docker-compose-db.yaml up -d`.
4.  Initialize databases with sample data (optional, but recommended).
5.  Start the application layer: `docker compose -f docker/docker-compose.yaml up -d --build`.

## 5. Success Criteria
*   All containers in both compose files are running and healthy.
*   The frontend is accessible via the forwarded port (e.g., 8080).
*   The backend API is accessible and can connect to the PostgreSQL and Redis databases.
