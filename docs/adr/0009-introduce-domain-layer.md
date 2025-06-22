# ADR-0009: Introduce simple Domain layer with entities and errors

## Status

- Date: 22.06.2025  
- Status: Accepted  
- Author: Artur Kliuchka  

## Context

Previously, each layer (repositories, services, handlers) defined its own version of domain entities (e.g. `Subscription`, `Weather`) and errors (`ErrInternal`, etc.).  
This led to duplication and required additional error mapping between layers.

## Decision

We introduced a lightweight `domain` package to centralize core business entities and errors:

```bash
internal/domain/
├── entities.go // Subscription, Weather, Frequency
└── errors.go // ErrInternal, ErrCityNotFound, ...
```

All layers now depend on a single shared definition of business data and common error values.

## Consequences

- Simplified error handling across layers (no more error remapping)
- Unified domain model across the application
- Slight increase in coupling, but justified by the project's scope
