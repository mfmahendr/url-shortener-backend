<div align="center">
    <h1>URL Shortener Backend</h1>
    <strong>A simple service for shortening URLs, tracking clicks, and managing domain restrictions.</strong>
    </br>
    </br>
    <div align="center">
        <span>
            <img src="https://img.shields.io/github/go-mod/go-version/mfmahendr/url-shortener-backend" alt="Go Version" />
            <img src="https://img.shields.io/github/license/mfmahendr/url-shortener-backend" alt="License" />
            <img src="https://img.shields.io/github/last-commit/mfmahendr/url-shortener-backend" alt="Last Commit" />
            <img src="https://img.shields.io/github/actions/workflow/status/mfmahendr/url-shortener-backend/google-cloudrun-docker.yml?branch=main" alt="CI Status" />
            <img src="https://img.shields.io/badge/API-live-blue" alt="Live API status" />
        </span>
    </div>
    <div align="center">
        <a href="https://shurl.my.id/" target="_blank"><strong>🔗 Visit the Live API</strong></a> ・
        <a href="http://docs.shurl.my.id/" target="_blank"><strong>📝 View Full API Docs</strong></a>
    </div>
    </br>
    <div align="center">
        <a href="#features">Features</a> - 
        <a href="#tech-stack">Tech Stack</a> - 
        <a href="#getting-started">Getting Started</a> - 
        <a href="#usage">Usage</a> - 
        <a href="#cicd-workflow">CI/CD</a> - 
        <a href="#license">License</a>
    </div>
</div>

<br>

This is an API allowing users to shorten long URLs and redirect from a short link to the original URL. Users can also track click analytics. Admins may manage blacklisted domains. The API is secured using Firebase Authentication and supports both public and authenticated endpoints.

## Features

* Create short links for a long URLs, with optional custom IDs
* Redirect to the original URL via short link
* Private shortlinks (only accessible by the creator)
* Track click analytics (IP address, user-agent, timestamp)
* Export click data in JSON or CSV format
* Domain blacklist support (admin only)
* Firebase JWT-based authentication for secure access
* Full OpenAPI 3.0 documentation

## Tech Stack

![Go](https://img.shields.io/badge/Go-1.23.1-blue?logo=go) ![Firebase Auth](https://img.shields.io/badge/Auth-Firebase-orange?logo=firebase) ![Firestore](https://img.shields.io/badge/Database-Firestore-ffca28?logo=firebase) ![Redis](https://img.shields.io/badge/Cache-Redis-dc382d?logo=redis) ![OpenAPI](https://img.shields.io/badge/OpenAPI-3.0-green?logo=swagger) ![Safe Browsing](https://img.shields.io/badge/Security-Google%20Safe%20Browsing-lightgrey?logo=google) ![Testing](https://img.shields.io/badge/Testing-Testify%20%7C%20Testcontainers%20%7C%20Redismock-blueviolet)

| Technology        | For what? |
|-------------------|-----------------|
| Go (Golang)       | Core backend language used to build the RESTful HTTP server. Uses `httprouter` for efficient routing. |
| REST Architecture | API follows REST principles (stateless, resource-oriented, standard methods). |
| Firebase Auth     | Handles user authentication via JWT tokens (email and password). |
| Firestore         | NoSQL database for storing shortlink data, user metadata, and click logs. |
| Redis             | Used for caching shortlink resolutions and rate limiting. Helps improve performance. |
| Safe Browsing API | Verifies if a submitted URL is malicious before shortening. |
| OpenAPI 3.0       | Specification for documenting the RESTful API. |
| Testify           | Testing framework for assertions in Go unit tests. |
| Testcontainers    | Spins up disposable containers (e.g. Redis and Firebase Emulator) for integration testing. |
| Redismock         | Mocks Redis connections for fast and isolated unit tests. |

## Getting Started

### Prerequisites
Make sure you have the following installed before running the projec

- Go 1.23.1 or newer version
- Firebase Project with service account key for local dev
- Safe Browsing API Key (from Google API Console)
- Redis (for local caching and rate limiting) or you may use a managed Redis service, such as Redis Cloud, Upstash, Google Cloud Memorystore, etc
- Docker  (for integration testing and optionally for containerized run)
- Google Cloud SDK (for manual deploy)

### Installation / Setup
1. Clone the repository

```bash
git clone https://github.com/mfmahendr/url-shortener-backend.git
cd url-shortener-backend
```
2. Download dependencies

```bash
go mod download
```

3. Build the application (if you want to) 

```bash
go build -o app ./cmd/app
```

### Environment Variables
Environment variables must be configured for both development and production. You can use a .env file for local setup. You can copy the `.env.example` file to `.env` and fill in the required values.

In production, these values should be passed as environment variables via your deployment tool (e.g. GitHub Actions + Cloud Run).
| Variable Name                  | Description                                                      |
|-------------------------------|------------------------------------------------------------------|
| `APP_ENV`                     | Application environment (must be `development` or `production`)       |
| `PORT`                        | Port number for the HTTP server (default: `8080`)                |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to your Firebase service account key JSON file (**local development only**). In production, use default credentials. |
| `FIREBASE_PROJECT_ID`         | Your Firebase project ID                                         |
| `REDIS_ADDR`                  | Redis server address (e.g. `localhost:6379`)                     |
| `REDIS_PASSWORD`              | Password for Redis instance                                      |
| `SAFE_BROWSING_API_KEY`       | Google Safe Browsing API key                                     |

### Run the Application
Locally using Go:
```bash
go run ./cmd/app
```

Or using Docker:

```bash
docker build -t url-shortener-backend .
docker run --env-file .env -e APP_ENV=development -p 8080:8080 url-shortener-backend
```

### Run Tests

**Unit Tests**

```bash
go test ./internal/...
```

**Integration Tests**

Requires Docker (used by testcontainers-go):

```bash
go test ./tests/integration/...
```

Integration test will automatically spin up containers for Firebase Emulator and Redis.





## Usage

The API exposes public and authenticated endpoints for creating and managing short URLs, including analytics and blacklist administration.

> Full API docs are available at:
> **[https://mfmahendr.github.io/url-shortener-backend/](https://mfmahendr.github.io/url-shortener-backend/)**
> *(Swagger UI served from GitHub Pages, powered by OpenAPI 3.0)*

#### Public Endpoints

* `GET /` → Welcome message
* `GET /health` → Health check
* `GET /r/{short_id}` → Redirect to the original URL (tracks click)

**Example:**  
To visit this GitHub repository via a short link, you may open:  
[https://shurl.my.id/r/mybackend](https://shurl.my.id/r/mybackend)

#### Authenticated Endpoints (`Authorization: Bearer <Firebase_JWT>`)

**URL Management:**

* `POST /u/shorten` → Create short URL (optional custom ID, support private links)
* `GET /u/click-count/{short_id}` → Get total clicks
* `GET /u/analytics/{short_id}` → Get click logs (with pagination + filters)
* `GET /u/click-count/{short_id}/export` → Export click logs (CSV/JSON)

**Admin Only:**

* `POST /admin/blacklist` → Add domain to blacklist
* `GET /admin/blacklist` → List all blacklisted domains
* `DELETE /admin/blacklist` → Remove domain from blacklist

For all available endpoints, request/response schema, and authorization rules, please refer to the [API documentation](https://mfmahendr.github.io/url-shortener-backend/).


## CI/CD Workflow

This project uses **GitHub Actions** for automated testing, building, and deployment. Here's how the pipeline works:

* **Trigger:**
  CI/CD runs on every push to the `main` branch, excluding changes to docs (`.md` files, `docs/`, etc.).

* **Test Stage:**
  Runs unit tests and integration tests.

* **Build Stage:**
  If all tests pass, the app is built into a Docker image and pushed to **Google Artifact Registry**. Authentication is handled using **Workload Identity Federation (WIF)**.

* **Deploy Stage:**
  The Docker image is deployed to **Google Cloud Run** using the `google-github-actions/deploy-cloudrun` action. Environment variables are injected securely using GitHub Secrets and Repository Variables.

* **Cloud Run Config:**

  * Service Name: `app`
  * Region: defined via `GCP_REGION` variable
  * Other environment variables is `FIREBASE_PROJECT_ID`, `REDIS_ADDR`, `REDIS_PASSWORD`, `SAFE_BROWSING_API_KEY`, etc.

## License

This project is licensed under the [MIT](https://choosealicense.com/licenses/mit/) License.
