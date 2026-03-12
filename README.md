# Credora Backend

## Overview

Credora is a backend system that simulates core **fintech / digital banking infrastructure**. The project models how modern financial systems process payments, manage accounts, ensure transactional consistency, and handle asynchronous financial events.

It is written in **Go** and designed using layered architecture with domain-driven concepts.

This project demonstrates how backend systems in fintech handle:

- Double-entry ledger accounting
- Idempotent payment processing
- Webhook event ingestion
- Asynchronous background workers
- Virtual account infrastructure
- Transaction settlement
- Extensive OpenAPI documentation

---

# Live Architecture

Here is a textual architecture diagram that illustrates the system design:

```
          +------------------+
          |      Client      |
          +------------------+
                   |
                   v
          +------------------+
          |  HTTP API (Chi)  |
          +------------------+
                   |
                   v
          +------------------+
          |   Service Layer  |
          +------------------+
                   |
                   v
          +------------------+
          | Repository Layer |
          +------------------+
                   |
                   v
+-----------------------+       +-------------------+
|    PostgreSQL DB      |       |  Redis Event Bus  |
+-----------------------+       +-------------------+
                                           |
                                           v
                                   +----------------+
                                   | Asynq Workers  |
                                   +----------------+
```

This diagram helps recruiters and engineers immediately understand how requests flow and how asynchronous processing is handled.

---

# Key Engineering Concepts Implemented

## Double-Entry Ledger System

Financial transactions are modeled using a double-entry accounting system.

Every transfer generates:

- a **debit ledger entry**
- a **credit ledger entry**

This guarantees balance consistency and mirrors how real banking ledgers operate.

## Idempotent Transfers

Transfers are designed to be **idempotent**, preventing duplicate financial transactions when clients retry requests.

## Idempotent Webhook Processing

External payment providers can retry webhook deliveries.

Credora ensures each webhook event is processed **exactly once** using idempotency tracking.

## Event-Driven Processing

The system publishes domain events to Redis which are processed asynchronously by background workers.

Examples:

- email notifications
- transfer processing
- account creation events
- sending noticiations via Firebase Cloud Messaging

---

# Core Features

## Authentication

- JWT authentication
- Account registration
- Login system

## Virtual Accounts

Integration with **Monnify** for reserved / virtual bank accounts.

Features:

- account reservation
- payment webhooks
- reconciliation

## Transaction Processing

- Internal transfers
- Ledger accounting
- Transaction history
- Settlement calculations

## Asynchronous Processing

Redis + Asynq workers power background jobs including:

- email notifications
- financial event handling

---

# API Documentation

```
/docs/swagger
```

or

```
http://localhost:8080/api/v1/documentation
```

---

# Transaction Flow

Here is a textual sequence diagram for internal transfers:

```
Client submits transfer request
        |
        v
API validates accounts
        |
        v
Create idempotency record
        |
        v
Create debit + credit ledger entries
        |
        v
Publish event to Redis
        |
        v
Worker processes notifications
```

<!-- ADD A MORE POLISHED DIAGRAM LATER -->

---

# Tech Stack

Language

- Go

Database

- PostgreSQL

Infrastructure

- Redis
- Firebase Cloud Messaging
- Mailtrap

Libraries

- Chi Router
- pgx
- SQLC
- Asynq

---

# Project Structure

```
cmd/
   server/
   worker/
   cli/

internal/
   config/
   router/
   handlers/
   queues/

service/
   business logic

infrastructure/
   repository implementations/ adapters to the 3rd parties like Firebase

 domain/
   models and shared utilities
```

---

# Running Locally

### Requirements

- Go 1.24+
- PostgreSQL
- Redis

### Setup

```
make migrate-up
make run
```

Start background worker

```
make start-worker
```

---

# Testing

Run tests:

```
make test
```

Integration tests:

```
make test-integration
```

---

# Future Improvements

⚠️ Add links to technical write-ups about the system

Example articles I could write:

- "Designing a Double-Entry Ledger in Go"
- "Handling Idempotent Payment Webhooks"
- "Building Event-Driven Systems with Redis and Go"

## Articles already published

- [The math behind scaling job processors](https://medium.com/@osamwonyiefosa02/understanding-queue-backlogs-worker-saturation-and-the-math-behind-scaling-job-processors-0136599eb48e)
- [Deploying a Dockerized Golang Server on EC2 and ECR](https://efosae.hashnode.dev/deploying-a-containerized-app-to-aws-ec2)
- [I Built My Own OpenAPI Generator in Go (Without a Single Library)](https://medium.com/@osamwonyiefosa02/i-built-my-own-openapi-generator-in-go-without-a-single-library-dabf21804794)

---

# Author

Efosa Osamwonyi

Backend Engineer

Focus areas:

- Backend systems
- Fintech infrastructure
- Event-driven architecture
