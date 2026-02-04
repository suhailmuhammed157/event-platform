# High-Performance Event Processing Platform (Go + Kafka)

A production-grade, high-throughput event processing platform built in Go.  
Designed to demonstrate system design, concurrency, reliability, and performance engineering.

## Goals

- Handle 100k+ events/sec on commodity hardware
- Bounded concurrency with backpressure
- At-least-once delivery with retry and DLQ
- Observability with Prometheus + Grafana
- Horizontally scalable via Kafka partitions

## Architecture

![Architecture](./architecture.png)

## Components

- **Ingest Service** – HTTP/gRPC API for event ingestion
- **Kafka** – Durable event transport
- **Processor Service** – Batch processing with worker pools
- **Sink Service** – Batched persistence to storage
- **Observability** – Prometheus + Grafana

## Delivery Guarantees

- At-least-once processing
- Idempotent sink writes
- Retry queues and DLQ

## Tech Stack

- Go
- Apache Kafka
- Prometheus + Grafana
- Docker Compose

## Local Setup

```bash
docker-compose up -d
```

## Project structure

event-platform/
├── ingest-service/
│ ├── cmd/ingest/main.go
│ ├── internal/
│ │ ├── http/
│ │ ├── config/
│ │ ├── batching/
│ │ └── producer/
│ └── Dockerfile
├── processor-service/
│ ├── cmd/processor/main.go
│ ├── internal/
│ │ ├── consumer/
│ │ ├── workers/
│ │ ├── retry/
│ │ └── dlq/
│ └── Dockerfile
├── sink-service/
│ ├── cmd/sink/main.go
│ └── internal/
│ └── storage/
├── pkg/
│ ├── event/
│ ├── config/
│ └── observability/
├── proto/
│ └── event.proto
├── loadgen/
│ └── main.go
├── docker-compose.yml
├── architecture.png
└── README.md
