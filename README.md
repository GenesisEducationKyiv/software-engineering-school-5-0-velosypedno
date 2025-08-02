# Weather Forecast API

Weather API application that allows users to subscribe to weather updates for their city.

## Table of Contents

- [Installation](#install)
- [Testing](#testing)
- [Setup Git hook](#setup-git-hook)
- [Documentation](#documentation)
- [License](#license)
- [Future Improvements](#improvements)

## Install

### Dependencies

Ensure you have the following installed:

- [Go-task](https://taskfile.dev/installation/)
- [Go](https://golang.org/doc/install) (>= 1.23.5)
- [Docker](https://docs.docker.com/get-docker/)

### Steps

0. **Check available tasks**:

    ```bash
    go-task
    ```

    *P.S. All tasks have description, so it is worth to check*

1. **Clone the repository**:

   ```bash
   git clone https://github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno.git
   ```

2. **Change work directory**:

    ```bash
    cd software-engineering-school-5-0-velosypedno
    ```

3. **Configure environmental variables**:

    Copy `.env.sample`

    ```bash
    go-task copy:env
    ```

    **NOTE**: `.env` must be edited manually. You need to set smtp credentials, API key, etc.

4. **Build and up services by Docker Compose**:

    ```bash
    go-task docker:up
    ```

    This will start the following services:
    - `postgres` – database
    - `redis` – cache
    - `rabbitmq` – message broker
    - `gateway` – unified HTTP entrypoint that routes requests to internal services
    - `weather` – gRPC service that fetches and provides current weather data
    - `sub` – core service (acts as a monolith): manages subscriptions, processes confirmation/unsubscription, and runs scheduled jobs
    - `notifier` – consumes events from RabbitMQ and delivers email notifications
    - `migrator` – one-time task that runs database schema migrations
    - `prometheus` & `grafana` – observability stack for metrics collection and visualization

    Check [`docker-compose.yml`](./docker-compose.yml) for more details

## Testing

- **To run all tests**:

    ```bash
    go-task test
    ```

- **Run unit tests**:

    ```bash
    go-task test:unit
    ```

- **Run integration tests**:

    ```bash
    go-task test:integration
    ```

    Integration tests use `.env.test`, start required Docker services, run migrations, then execute tests with `integration` tag.

## Setup Git hook

- **To install the pre-commit Git hook that runs linter automatically before each commit:**

    ```bash
    go-task copy:hooks:pre-commit
    ```

- **To remove the pre-commit Git hook:**

    ```bash
    go-task rm:hooks:pre-commit
    ```

## Documentation

- Detailed system design documents, ADRs, and diagrams are available in the [`./docs`](./docs/) folder.
- System Design Document - [here](./docs/sdd/document.md)
- Swagger Scheme - [here](./docs/sdd/swagger.yaml)

## Improvements

### Alerts and Monitoring

To ensure system reliability and early detection of critical issues, we suggest configuring a minimal yet effective set of alerts based on logs and metrics.

#### Log-based Alerts

These alerts rely on structured logs and their severity levels:

- **High error log rate**: Alert when the number of logs with `ERROR` level exceeds a threshold (e.g., 10 per minute). In this system, error logs indicate critical failures that require immediate attention.
- **Subscription flow failures**: Alert when an internal error prevents subscription creation, activation, or deletion. This may signal persistent service or DB issues.

#### Metrics-based Alerts

These alerts use Prometheus metrics collected via instrumented middleware:

- **High client/server error ratio**: Trigger an alert when the sum of HTTP 4xx and 5xx responses exceeds 50% of total requests over a given window. This may indicate bad API usage or internal malfunction.
- **Low notification delivery rate**: Alert when the percentage of successfully sent notifications drops below 50% of total attempts. Undelivered messages are re-queued, indicating persistent delivery failures.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details
