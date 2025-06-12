# ADR-0003: Use separate migrator service

## Status

- Date: 08.06.2025
- Status: Accepted
- Author: Artur Kliuchka

## Context

The application requires database migrations to be executed before the backend services (`api-weather`, `api-cron`) start using the database schema

There were two options:

1. Run migrations inside the API and Cron services on startup
2. Create a separate service dedicated to running migrations

## Decision

We created a standalone service `migrator` that runs database migrations using the `golang-migrate` CLI tool.  
See [`docker-compose.yml`](/docker-compose.yml), where the service is defined.

## Consequences

- Separation of concerns: API/Cron services do not contain migration logic
- More flexible deployment and control over migrations
- Easy to rerun or test migrations in isolation
