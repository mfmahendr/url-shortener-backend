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
    <strong>Backend service for shortening URLs, tracking clicks, and managing domain restrictions.</strong>
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

| Technology        | For what exactly? |
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

## API Documentation

TODO

## License

This project is licensed under the MIT License.
