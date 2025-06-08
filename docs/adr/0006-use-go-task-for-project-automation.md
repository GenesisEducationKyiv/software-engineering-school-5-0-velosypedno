# ADR-0006: Use go-task for project automation

## Status

- Date: 08.06.2025  
- Status: Accepted  
- Author: Artur Kliuchka

## Context

We need a consistent and cross-platform way to automate common development and CI tasks (e.g. linting, migrations, tests, service startup).  
Two options were considered:

1. Use `make` with a `Makefile`
2. Use [`go-task`](https://taskfile.dev) with a `Taskfile.yml`

## Decision

We chose `go-task` because:

- It has native support for YAML syntax, which improves readability
- It is cross-platform (unlike `make`, which has poor support on Windows)
- It integrates well with Go projects
- It supports task composition and includes, useful for managing multiple environments (e.g. Docker)

`Taskfile.yml` is used locally and in CI to manage linting, tests, Docker, and database migration workflows.

## Consequences

- Developer experience is improved through clear and self-documented tasks
- CI pipelines can reuse the same logic as local development
- Easier to onboard new developers by exposing all commands via `go-task --list`
