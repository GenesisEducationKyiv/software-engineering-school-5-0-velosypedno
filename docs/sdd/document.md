# System Design: Weather Subscription Service

## 1. System Requirements

### Functional

- Users can subscribe to weather updates for a specific city
- The system should send notifications to users' emails (hourly, daily)
- The user should be able to unsubscribe
- Users should be able to subscribe to a weekly weather forecast.
- In case of problems, the user should be able to contact support

### Non-Functional

- **Reliability**: emails should be delivered within 3 minutes of the specified time
- **Completeness**: If the system fails and a weather update message cannot be delivered, notify the user.
- **Resiliency**: in the event of a complete failure, the system must fully recover and perform a requeue
- **Security**: the system must withstand various types of cyber attacks
- **Auditability**: the system thoroughly collects all errors and logs for further manual analysis, or for use in analytics

### Constraints

- External weather API limit: 1.000.000 per month
- Budget: 0$ (pet project)

## 2. System load

- Active users < 20
- Subscriptions per user 1-10

## 3. High-level Architecture

![img.png](../adr/0010-microservices.png)

Check [`Microservices ADR`](../adr/0010-monolit-to-microservices.md)

## 4. Microservices' Architecture Diagrams

## Gateway

![img.png](../../gateway/docs/application-architecture.png)

## API endpoints

All routes are prefixed with `/api`.

| Method | Endpoint              | Description                                                                |
|--------|-----------------------|----------------------------------------------------------------------------|
| GET    | `/weather`            | Get current weather for a given city. Requires `?city=CityName` query.     |
| POST   | `/subscribe`          | Subscribe a user to weather updates. Expects JSON body with email, city, and frequency (`hourly` or `daily`). |
| GET    | `/confirm/:token`     | Confirm a subscription via token received by email.                        |
| GET    | `/unsubscribe/:token` | Unsubscribe from weather notifications using the token.                    |

## Weather

![img.png](../../weather/docs/application-architecture.png)

## Subscription

![img.png](../../sub/docs/application-architecture.png)

### DB schema

![img.png](./db-relations.png)

## Notifier

![img.png](../../notifier/docs/application-architecture.png)