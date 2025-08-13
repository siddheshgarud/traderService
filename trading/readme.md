# Trading Exchange Events Producer & Processor

## Overview

This project consists of two microservices:

- **Producer Service:** Publishes trading events (orders, trades, cash balances) to Kafka topics.
- **Processor Service:** Consumes these events from Kafka and processes or prints them.

The system enables real-time event-driven communication between trading components.

---

## Kafka Topics & Schemas

### 1. Orders

**Topic:** `orders`

| Field     | Description         |
|-----------|--------------------|
| client_id | Unique client ID   |
| order_id  | Unique order ID    |
| qty       | Order quantity     |
| price     | Order price        |
| remark    | Additional remarks |

### 2. Trades

**Topic:** `trades`

| Field             | Description                  |
|-------------------|-----------------------------|
| client_id         | Unique client ID            |
| order_id          | Associated order ID         |
| qty               | Trade quantity              |
| price             | Trade price                 |
| remark            | Additional remarks          |
| trade_id          | Unique trade ID             |
| partial_filled_qty| Quantity partially filled   |

### 3. Cash Balance

**Topic:** `cash_balance`

| Field           | Description                  |
|-----------------|-----------------------------|
| client_id       | Unique client ID            |
| pay_in          | Amount paid in              |
| payout          | Amount paid out             |
| opening_balance | Opening balance             |
| margin_utilized | Margin utilized             |

---

## Event Flow

1. **Producer Service** constructs and sends messages to the respective Kafka topics.
2. **Processor Service** subscribes to these topics, consumes events, and prints or processes them.

---

## Scaling Guidelines

### Scaling the Producer

- Deploy multiple instances of the producer service to increase throughput and parallelism.
- Ensure Kafka topics have sufficient partitions to distribute load evenly.
- Use message keys to control partition assignment when message ordering is required.
- Each producer instance independently sends messages to Kafka brokers, allowing horizontal scaling.
- Monitor resource usage and adjust the number of producer instances according to load.

### Scaling the Consumer

- Increase the number of partitions in Kafka topics to enable concurrent processing.
- Deploy multiple consumer instances configured with the same consumer group ID.
- Kafka distributes partitions among consumer instances within the same group, enabling parallel consumption.
- When a consumer instance stops, Kafka rebalances partitions to remaining consumers.
- Scale consumers by adding more instances up to the total number of partitions.
- Ensure consumer message handlers are stateless or correctly manage state to avoid processing conflicts.

---

## Getting Started

1. **Start Kafka and Zookeeper** (if not already running).
2. **Build and run the producer and processor services** using Docker Compose or your preferred method.
3. **Monitor logs** to verify events are being produced and consumed as expected.

---

## Best Practices

- Use environment variables for configuration (e.g., Kafka broker addresses).
- Implement error handling and retries in both producer and consumer.
- Monitor Kafka topic lag and service health.
- Secure sensitive data and communication channels.

---
