# ADR-0001: Use postgres database

## Status

- Date: 08.06.2025
- Status: Accepted
- Author: Artur Kliuchka

## Context

The project need a relational database to store structured data.\
We require basic ACID properties, schema enforcement and support SQL.

## Decision

We chose `PostgreSQL` over `SQLite` for the following reasons:

- PostgreSQL can support concurrent access and is production-ready
- SQlite is file-based and nit suitable for production
- PostgreSQL is widely used and actively maintained

## Consequences

We run PostgreSQL in docker container (`postgres-weather`)
