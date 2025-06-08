# ADR-0001: Use postgres database

## Status

- Date: 08.06.2025
- Status: Accepted
- Author: Artur Kliuchka

## Context

The project needs a relational database to store structured data.  
We require basic ACID properties, schema enforcement and SQL support.

## Decision

We chose `PostgreSQL` over `SQLite` for the following reasons:

- PostgreSQL supports concurrent access and is production-ready
- SQlite is file-based and not suitable for production
- PostgreSQL is widely used and actively maintained

## Consequences

We run PostgreSQL in docker container (`postgres-weather`).  
See [`docker-compose.yml`](/docker-compose.yml), where service is defined.
