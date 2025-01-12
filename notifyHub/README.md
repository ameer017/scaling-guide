# NotifyHub

A standalone microservice for sending **email** notifications — built primarily as a hands-on project to **learn and understand Kafka**.

The service is the vehicle; Kafka is the focus. By building a real producer/consumer pipeline (API publishes jobs, a worker consumes them, retries and failure handling ride on top of Kafka), you'll get practical experience with topics, consumer groups, offsets, delivery guarantees, and how async messaging fits into a microservice.

Emails are accepted over a REST API, persisted as `PENDING`, published to **Kafka**, and delivered asynchronously by a worker with retries, delivery logging, and optional scheduling/templates.

## Tech Stack

| Area           | Choice                   |
| -------------- | ------------------------ |
| Language       | Go                       |
| HTTP router    | Chi                      |
| Database       | PostgreSQL               |
| Cache          | Redis                    |
| Message broker | **Kafka**                |
| Containers     | Docker + Docker Compose  |
| Auth           | JWT (service-to-service) |
| Logging        | Zerolog or slog          |
| Data access    | SQLC or GORM             |
| API docs       | Swagger / OpenAPI        |

## Suggested Folder Structure

```text
notifyHub/
├── cmd/
│   ├── api/                 # REST API entrypoint
│   └── worker/              # Kafka consumer / delivery worker
├── internal/
│   ├── config/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── middleware/
│   ├── models/
│   ├── queue/               # Kafka producer / consumer wrappers
│   ├── email/
│   └── scheduler/
├── pkg/
│   ├── logger/
│   └── validator/
├── migrations/
├── docs/                    # OpenAPI / Swagger
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

## Database Design

### `notifications`

| Column         | Description                          |
| -------------- | ------------------------------------ |
| `id`           | Primary key                          |
| `recipient`    | Destination email address            |
| `subject`      | Email subject                        |
| `body`         | Email body (plain text or HTML)      |
| `status`       | `PENDING` \| `SENT` \| `FAILED` \| … |
| `scheduled_at` | When to send (null = send ASAP)      |
| `sent_at`      | When delivery succeeded              |
| `created_at`   | Created timestamp                    |
| `updated_at`   | Last update timestamp                |

### `templates`

| Column       | Description       |
| ------------ | ----------------- |
| `id`         | Primary key       |
| `name`       | Template name     |
| `subject`    | Subject template  |
| `body`       | Body template     |
| `created_at` | Created timestamp |

### `delivery_logs`

| Column            | Description                                     |
| ----------------- | ----------------------------------------------- |
| `id`              | Primary key                                     |
| `notification_id` | FK → `notifications`                            |
| `provider`        | Provider used (e.g. Mailtrap, Resend, SendGrid) |
| `response`        | Provider response / error payload               |
| `status`          | Attempt outcome                                 |
| `attempt`         | Attempt number                                  |
| `created_at`      | Created timestamp                               |

## API Endpoints

| Method   | Path                  | Description                            |
| -------- | --------------------- | -------------------------------------- |
| `POST`   | `/notifications`      | Create / enqueue an email notification |
| `GET`    | `/notifications/{id}` | Get one notification                   |
| `GET`    | `/notifications/{id}/logs` | List delivery attempts            |
| `GET`    | `/notifications`      | List notifications                     |
| `DELETE` | `/notifications/{id}` | Cancel / delete a notification         |
| `POST`   | `/templates`          | Create a template                      |
| `GET`    | `/templates`          | List templates                         |
| `GET`    | `/health`             | Health check                           |
| `GET`    | `/metrics`            | Metrics                                |

## Learning Goals (Kafka)

This project is intentionally shaped around Kafka so you can practice:

- Producing messages from the API after persisting a notification
- Consuming with a worker (consumer groups, offsets, rebalancing)
- Choosing topic design (e.g. `notifications.send`) and message payloads
- Retries, backoff, and dead-letter topics when delivery fails
- How Kafka sits between HTTP, Postgres, and side-effectful workers

Email delivery is the domain; Kafka is what we're here to understand.

## Queue Flow (Kafka)

1. API receives an email notification request.
2. Save it as `PENDING` in PostgreSQL.
3. Publish the notification **ID** to a Kafka topic (e.g. `notifications.send`).
4. Worker consumes the message from Kafka.
5. Send the email via the configured provider.
6. Update the notification status to `SENT` or `FAILED`.
7. Retry failures with exponential backoff (and optional delayed re-publish to Kafka).
8. Mark permanently failed after a configurable max number of attempts; write each attempt to `delivery_logs`.

```text
Client → API → PostgreSQL (PENDING)
              → Kafka topic
                    ↓
                 Worker → Email provider → status SENT / FAILED
                                         → delivery_logs
```

## Email Provider

**Gmail SMTP** (Go equivalent of Nodemailer + Gmail):

1. Enable 2-Step Verification on the Google account.
2. Create an [App Password](https://myaccount.google.com/apppasswords).
3. Set these in `.env`:

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=you@gmail.com
SMTP_PASSWORD=your-16-char-app-password
SMTP_FROM=you@gmail.com
```

The worker sends mail over SMTP with STARTTLS.

### Retries, delivery logs, and DLQ

- Each send attempt is written to `delivery_logs` (`SUCCESS` / `FAILURE`).
- Failed attempts retry with exponential backoff (`RETRY_BACKOFF_SECONDS`, default 2s → 2s, 4s, 8s…) up to `MAX_DELIVERY_ATTEMPTS` (default 3).
- After retries are exhausted, status becomes `FAILED` and a message is published to `notifications.dlq`.
- Inspect attempts: `GET /notifications/{id}/logs`
- Inspect DLQ: `make kafka-dlq`

## Future Enhancements

- SMS and push channels
- Webhooks for delivery updates
- User notification preferences
- Rate limiting per recipient
- Batch notifications
- Dead-letter topic (Kafka DLQ) — **done** (`notifications.dlq`)
- Multi-tenancy
- Event sourcing
- gRPC API
- Kubernetes deployment
- Prometheus metrics
- OpenTelemetry tracing

## Development Plan

1. Set up the Go project, Docker, PostgreSQL, Kafka, and Redis — get comfortable inspecting topics and consumer groups locally.
2. Build the REST API to create and query notifications.
3. Implement Kafka publishing and a worker to consume jobs (core learning step).
4. Integrate an email provider and send real emails.
5. Add retries, logging, delivery tracking, and a dead-letter topic.
6. Add templates and scheduled notifications.
7. Write tests and add API documentation.

## Local Development

```bash
# from notifyHub/
cp .env.example .env    # optional; defaults already match compose

make up                 # Postgres, Redis, Kafka + create notifications.send topic
make migrate            # create notifications table
make kafka-topics       # should list notifications.send

make run-api            # http://localhost:8080/health
make run-worker         # consumes notifications.send and sends via Gmail SMTP

# Enqueue a notification (API persists PENDING, then publishes ID to Kafka)
curl -s -X POST http://localhost:8080/notifications \
  -H 'Content-Type: application/json' \
  -d '{"recipient":"you@example.com","subject":"Hello","body":"Kafka learning"}'

# Inspect consumer group lag / offsets
make kafka-describe-group

make down               # stop infrastructure
```

Useful Kafka targets: `make kafka-topics`, `make kafka-groups`, `make kafka-describe-group`.
