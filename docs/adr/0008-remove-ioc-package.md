# ADR-0007: Replace `ioc` package with idiomatic `app` package

## Status

- Date: 22.06.2025  
- Status: Accepted  
- Author: Artur Kliuchka

## Context

Previously, the application used a custom package `internal/ioc` (see [ioc](./0005-ioc-containers.md)) to handle dependency injection for both the API server and scheduled tasks.  
While this centralized the creation of services and handlers, it introduced a non-idiomatic abstraction that doesn't align with typical Go practices.

Go projects usually don't use an "IoC container" approach like in other languages (e.g., Java Spring or C#).  
Instead, Go favors **explicit wiring**, often done in the `main.go` or a dedicated `app` package that orchestrates the composition of components.

## Decision

We removed the `ioc` package and replaced it with an idiomatic `internal/app` package that:

- Encapsulates the application lifecycle (startup, shutdown)
- Initializes components explicitly (database, services, API server, cron scheduler)
- Provides clear entry points for different parts of the system

## Consequences

- More idiomatic structure for Go applications
- Improved clarity and separation of concerns
- Easier to manage lifecycle events (e.g. graceful shutdown)
- Slightly more verbose setup code compared to centralized ioc
