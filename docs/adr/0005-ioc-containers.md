# ADR-0005: Use internal/ioc for dependency injection

## Status

- Date: 08.06.2025  
- Status: Accepted  
- Author: Artur Kliuchka  

## Context

The application relies on several interconnected components (repositories, services, tasks).  
Manually wiring them across entry points (API, Cron) would lead to duplication and tight coupling.

## Decision

We created a dedicated package `internal/ioc` to centralize dependency injection and initialization.

- `ioc/handlers.go`: builds and injects HTTP handler dependencies for the API.
- `ioc/tasks.go`: builds and injects task functions for the Cron service.

This approach helps keep entry points clean and separates composition from logic.

## Consequences

- Centralized construction of dependencies
- Easier testing and refactoring
- Clear overview of application wiring
