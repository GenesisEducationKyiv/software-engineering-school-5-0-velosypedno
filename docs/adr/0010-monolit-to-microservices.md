# ADR-0010: Microservices Communication Strategy — Mixed Approach (gRPC & Message Broker) with Centralized Subscription Service

## Status

- Date: 07.07.2025
- Status: Accepted (Updated)
- Author: Artur Kliuchka

## Context

We are splitting our monolithic backend into microservices:

![img.png](./0010-microservices.png)

- Subscription Service
- Weather Service
- Mailer Service

The **Subscription Service** is now considered the core of the application, managing its own database and orchestrating key business logic, including the scheduling and initiation of email dispatches. The **Weather Service** is responsible solely for providing weather data with its internal mechanisms (caching, fallback, circuit breakers), and the **Mailer Service** is a dedicated utility for sending emails as instructed.

A key decision is how these services should communicate. We considered and adopted a mixed communication strategy based on the nature of the interaction.

## Decision

We chose a mixed approach for inter-service communication, leveraging both gRPC and Message Broker:

1.  **gRPC for Synchronous, High-Performance Communication:**
    * **User to Business Services (via HTTP Gateway):** Direct communication from the HTTP Gateway to the Subscription Service and Weather Service uses gRPC for its efficiency and strong typing. This ensures low-latency responses for user-initiated requests.
    * **Internal Service-to-Service (Subscription Service to Weather Service):** When the Subscription Service needs current weather data, it makes a direct gRPC call to the Weather Service. This ensures efficient, real-time data exchange where an immediate response is required.

2.  **Message Broker for Asynchronous, Reliable Email Communication (initiated by Subscription Service):**
    * The **Subscription Service** is responsible for initiating email dispatches. It utilizes a scheduled task ("хрон таска") to process its data and generate email requests.
    * For sending these emails, the Subscription Service sends a message to the Message Broker. The Mailer Service then consumes these messages asynchronously. This provides reliable, asynchronous message delivery, buffering messages if the Mailer Service is temporarily unavailable, and supports high throughput for email notifications.
    * This approach ensures loose coupling between the Subscription Service and the Mailer, allowing the Mailer to operate independently and scale separately.

While gRPC offers low latency synchronous calls, it lacks built-in retry and buffering for scenarios where the receiving service might be temporarily unavailable. For critical asynchronous events like email notifications, where delivery guarantee and resilience to service downtime are essential, the Message Broker is more suitable. Conversely, for direct, real-time interactions where an immediate response is expected, gRPC is preferred.

## Consequences

-   **Subscription Service as Core:** The Subscription Service becomes the central hub for user subscriptions and related business logic, including data persistence and email dispatch orchestration.
-   **Clearer Service Specialization:** Weather Service remains a specialized data provider, and Mailer Service remains a specialized email sender, simplifying their internal logic and deployment.
-   **Optimized Communication:** We achieve optimized communication by choosing the most appropriate protocol for each interaction type: gRPC for performance-critical synchronous calls and Message Broker for resilient asynchronous email events.
-   **Single Database Ownership (Subscription Service):** Only the Subscription Service will manage its own database, simplifying data consistency concerns across other services.
-   **Mailer Service consumes email events asynchronously for scalability and reliability.**
-   **Increased Complexity:** Managing two distinct communication patterns (gRPC and Message Broker) adds a layer of architectural complexity compared to a single-pattern approach.
-   **Observability:** We will leverage Prometheus and Grafana for monitoring metrics from all business services and the HTTP Gateway, ensuring comprehensive observability across both synchronous and asynchronous communication paths.