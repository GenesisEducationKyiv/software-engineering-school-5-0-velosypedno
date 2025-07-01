# ADR-0002: Split backend into API and Cron services

## Status

- Date: 08.06.2025
- Status: Accepted
- Author: Artur Kliuchka

## Context

The backend application has two main responsibilities:

- serving HTTP API requests
- executing scheduled background tasks (e.g., sending emails)

## Decision

We split the backend into two independent services  

- `api-weather` - API service fow weather subscription
- `cron-weather` - run periodic background tasks

Each service run in its own Docker container.  
See [`docker-compose.yml`](/docker-compose.yml), where both services are defined.

## Consequences

- Services can be scaled, deployed and restarted independently
- Better separation of concerns
