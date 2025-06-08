# ADR-0004: Split backend into Services, Handlers, and Repositories

## Status

- Date: 08.06.2025  
- Status: Accepted  
- Author: Artur Kliuchka

## Context

The backend needs a clear and maintainable structure to support features like weather subscriptions, background jobs, and database access.

To keep the codebase modular and testable, we must separate concerns between:

- HTTP request processing  
- business logic  
- data access

## Decision

We adopted the following architectural structure:

- **Handlers** handle HTTP requests, extract parameters, call services, and return responses.
- **Services** contain business logic, validation, and orchestration of data access.
- **Repositories** are responsible for interacting with the database.

Each layer depends only on the layer directly below it (Handler → Service → Repository), making the system loosely coupled and easier to test.

## Consequences

- Improves readability and maintainability
- Simplifies unit testing by mocking dependencies
- Allows business logic reuse across different interfaces (e.g., HTTP, CLI, Cron)
