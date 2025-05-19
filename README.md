# Weather Forecast API

Weather API application that allows users to subscribe to weather updates for their city.

## Table of Contents

- [Installation](#install)
- [Testing](#testing)
- [API](#api)
- [Architecture](#architecture)
- [License](#license)

## Install

### Dependencies

Ensure you have the following installed:

- [Go](https://golang.org/doc/install) (>= 1.23.5)
- [Docker](https://docs.docker.com/get-docker/)

### Steps

1. **Clone the repository**:

   ```bash
   git clone https://github.com/velosypedno/genesis-weather-api.git
   ```

2. **Change work directory**:

    ```bash
    cd genesis-weather-api
    ```

3. **Configure environmental variables**:

    Copy `.env.sample`

    ```bash
    cp .env.sample .env
    ```

    **NOTE**: `.env` must be edited manually. You need to set smtp credentials, API key, etc.

4. **Build and up services by Docker Compose**:

    ```bash
    docker compose up
    ```

    This will start the following services:
    - `postgres-wether` - container with postgres database
    - `migrator` - waits for the database to start and then runs the migrations
    - `api-weather` - starts after the `migrator` finishes working, contains API
    - `cron-wether` - starts after the `migrator` finishes working, contains cron tasks to send email

## Testing

- **To run all tests**:

    ```bash
    go test -v ./...
    ```

## API

All routes are prefixed with `/api`.

| Method | Endpoint              | Description                                                                |
|--------|-----------------------|----------------------------------------------------------------------------|
| GET    | `/weather`            | Get current weather for a given city. Requires `?city=CityName` query.     |
| POST   | `/subscribe`          | Subscribe a user to weather updates. Expects JSON body with email, city, and frequency (`hourly` or `daily`). |
| GET    | `/confirm/:token`     | Confirm a subscription via token received by email.                        |
| GET    | `/unsubscribe/:token` | Unsubscribe from weather notifications using the token.                    |

## Architecture

This project follows layered architecture with a clear division of responsibilities. The structure is organized into the following layers:

- **Handlers** – handle HTTP requests, validate input, and return responses.
- **Services** – contain business logic (e.g., subscriptions, confirmation, weather processing).
- **Repositories** – provide access to PostgreSQL and external APIs.

```plaintext
.
├── cmd/               
│   ├── api/            # Main HTTP server startup
│   └── cron/           # Scheduled tasks for sending weather emails
└── internal/
    ├── config/          
    ├── handlers/       # HTTP requests handlers
    ├── ioc/            # Dependency injection 
    ├── models/         
    ├── repos/          # Repositories
    ├── scheduler/      # Cron tasks setup
    ├── server/         # HTTP server setup
    └── services/       # Business logic layer
```

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details
