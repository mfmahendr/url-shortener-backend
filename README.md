
<div align="center">
    <h1>URL Shortener Backend</h1>
    <!-- <div align="center">
        <img src="https://img.shields.io/github/go-mod/go-version/mfmahendr/url-shortener-backend" />
        <img src="https://img.shields.io/github/license/mfmahendr/url-shortener-backend" />
        <img src="https://img.shields.io/github/actions/workflow/status/mfmahendr/url-shortener-backend/ci.yml?branch=main" />
        <img src="https://img.shields.io/codecov/c/github/mfmahendr/url-shortener-backend" />
        <img src="https://img.shields.io/github/last-commit/mfmahendr/url-shortener-backend" />
    </div> -->
    </br>
    <strong>A simple service for shortening URLs, tracking clicks, and managing domain restrictions.</strong>
    <div align="center">
        <a href="#features">Features</a> - 
        <a href="#tech-stack">Tech Stack</a> - 
        <a href="#getting-started">Getting Started</a> - 
        <a href="#usage">Usage</a> - 
        <a href="#license">License</a>
    </div>
</div>


<br>

This is an API allowing users to shorten a long URLs and redirect from a short link to the original URL. User can also track click count from the short URL created. An admin may manage some domain to be blacklisted. The API is secured using Firebase Authentication and supports both public and authenticated endpoints.

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

TODO

## Usage

#### Get welcome message

```http
GET /
```

Returns a welcome message to verify the API is accessible.

| Response | Content                                             |
| -------- | --------------------------------------------------- |
| `200`    | `{ "message": "Welcome to the URL Shortener API" }` |
| `500`    | Internal server error                               |

---

#### Health check

```http
GET /health
```

Checks if the server is up and running.

| Response | Content                                              |
| -------- | ---------------------------------------------------- |
| `200`    | `{ "status": "ok", "message": "Server is running" }` |
| `500`    | Internal server error                                |

---

#### Redirect to original URL

```http
GET /r/{short_id}
```

Redirects to the original URL. Click tracking is performed.

| Parameter  | Type     | Description                |
| ---------- | -------- | -------------------------- |
| `short_id` | `string` | **Required.** Short URL ID |

| Response | Description                          |
| -------- | ------------------------------------ |
| `302`    | Redirects to the original URL        |
| `400`    | Bad request                          |
| `403`    | Access denied (private or forbidden) |
| `404`    | Short ID not found                   |
| `500`    | Server error                         |

---

#### Create a shortened URL

```http
POST /u/shorten
```

Authenticated endpoint to shorten a URL.

| Body Parameter | Type      | Description                              |
| -------------- | --------- | ---------------------------------------- |
| `url`          | `string`  | **Required.** Long URL to shorten        |
| `custom_id`    | `string`  | Optional custom short ID                 |
| `is_private`   | `boolean` | Optional flag to mark the URL as private |

| Response | Content                     |
| -------- | --------------------------- |
| `200`    | `{ "short_id": "abc123" }`  |
| `400`    | Invalid input               |
| `401`    | Unauthorized                |
| `403`    | Forbidden (e.g. unsafe URL) |
| `409`    | Custom ID conflict          |
| `500`    | Server error                |

---

#### Get total click count

```http
GET /u/click-count/{short_id}
```

Returns total clicks for a shortlink owned by the user.

| Parameter  | Type     | Description                |
| ---------- | -------- | -------------------------- |
| `short_id` | `string` | **Required.** Short URL ID |

| Response | Content                                       |
| -------- | --------------------------------------------- |
| `200`    | `{ "short_id": "abc123", "click_count": 42 }` |
| `400`    | Bad request                                   |
| `401`    | Unauthorized                                  |
| `403`    | Forbidden                                     |
| `404`    | Not found                                     |
| `500`    | Server error                                  |

---

#### Export click data

```http
GET /u/click-count/{short_id}/export
```

Exports click logs as CSV or JSON.

| Parameter  | Type     | Description                         |
| ---------- | -------- | ----------------------------------- |
| `short_id` | `string` | **Required.** Short URL ID          |
| `format`   | `string` | Optional. `csv` (default) or `json` |

| Response | Content            |
| -------- | ------------------ |
| `200`    | CSV or JSON export |
| `400`    | Bad request        |
| `401`    | Unauthorized       |
| `403`    | Forbidden          |
| `415`    | Unsupported format |
| `500`    | Server error       |

---

#### Get analytics

```http
GET /u/analytics/{short_id}
```

Returns click analytics with pagination and filters.

| Parameter    | Type      | Description                           |
| ------------ | --------- | ------------------------------------- |
| `short_id`   | `string`  | **Required.** Short URL ID            |
| `limit`      | `integer` | Max records to return                 |
| `cursor`     | `string`  | Pagination cursor (RFC3339 timestamp) |
| `after`      | `string`  | Filter clicks after this datetime     |
| `before`     | `string`  | Filter clicks before this datetime    |
| `order_desc` | `boolean` | Sort results in descending order      |

| Response | Content                               |
| -------- | ------------------------------------- |
| `200`    | JSON with analytics data and metadata |
| `400`    | Bad request                           |
| `401`    | Unauthorized                          |
| `403`    | Forbidden                             |
| `404`    | Not found                             |
| `500`    | Server error                          |

---

#### Add domain to blacklist (Admin only)

```http
POST /admin/blacklist
```

Adds a domain to the blacklist.

| Body Parameter | Type     | Description                   |
| -------------- | -------- | ----------------------------- |
| `domain`       | `string` | **Required.** Domain to block |

| Response | Content                                  |
| -------- | ---------------------------------------- |
| `200`    | `{ "status": "added", "domain": "..." }` |
| `400`    | Invalid input                            |
| `401`    | Unauthorized                             |
| `403`    | Forbidden                                |
| `409`    | Domain already blacklisted               |
| `500`    | Server error                             |

---

#### Get blacklist (Admin only)

```http
GET /admin/blacklist
```

Returns all blacklisted domains.

| Response | Content                 |
| -------- | ----------------------- |
| `200`    | Array of domain strings |
| `401`    | Unauthorized            |
| `403`    | Forbidden               |
| `500`    | Server error            |

---

#### Remove domain from blacklist (Admin only)

```http
DELETE /admin/blacklist/{domain}
```

| Parameter | Type     | Description                    |
| --------- | -------- | ------------------------------ |
| `domain`  | `string` | **Required.** Domain to remove |

| Response | Content                                    |
| -------- | ------------------------------------------ |
| `200`    | `{ "status": "removed", "domain": "..." }` |
| `400`    | Bad request                                |
| `401`    | Unauthorized                               |
| `403`    | Forbidden                                  |
| `404`    | Not found                                  |
| `500`    | Server error                               |

---


## License

This project is licensed under the [MIT](https://choosealicense.com/licenses/mit/) License.
