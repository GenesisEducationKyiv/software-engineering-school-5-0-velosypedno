# ADR-0007: Merge Cron service into API service as a goroutine

## Status

- Date: 22.06.2025  
- Status: Accepted  
- Author: Artur Kliuchka

## Context

Previously, the backend was split into two services:  
See [`ADR2`](./0002-split-backed-into-sevices.md)

- `api-weather` serving HTTP API requests  
- `cron-weather` running scheduled background tasks (cron jobs)  

Both services shared the same Dockerfile and accessed the same database. This split was recognized as an anti-pattern rather than a true microservice architecture because:

- The services tightly coupled through a shared database  
- Deployment and scaling complexity increased unnecessarily  
- Both services duplicated much of the same environment and configuration  

## Decision

We decided to remove the separate `cron-weather` service and instead run all scheduled background tasks as goroutines within the existing `api-weather` service process.  

This means:

- No separate Docker container for the cron service  
- Cron tasks run as scheduled goroutines inside the API service  
- Shared lifecycle and configuration with the API service  
- Database access is unified and simplified  

## Consequences

- Simplifies deployment by having only one backend service  
- Eliminates redundant containers and duplicated environment  
- Removes anti-pattern of sharing the same database between multiple services  
- Limits the ability to independently start/stop or scale the cron tasks separately  
- Reduces operational overhead and resource usage  
