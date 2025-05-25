# iDTask

A simplified high-concurrency distributed task scheduler built with Go and Python, demonstrating microservices architecture, intelligent scheduling, and observability. 

## 🔧 Tech Stack

* **Go**: Core scheduler, API gateway, and workers
* **Python**: AI-based task prediction and scheduling
* **Redis**: Job queue and fast cache
* **PostgreSQL**: Task and log storage
* **Next.js + Tailwind + Shadcn/UI**: Frontend for task monitoring
* **Prometheus**: Monitoring and metrics
* **gRPC**: Inter-service communication between Go and Python services
* **etcd**: Distributed coordination and leader election
* **Docker Compose**: Local orchestration of services

## 📐 System Architecture

```mermaid
graph LR
  A[Web Frontend - Next.js] --> B[API Gateway - Go]
  B --> C[Scheduler Service - Go and Python]
  C --> D[Worker Nodes - Go]
  C --> E[AI Module - Python ML]
  C --> F[PostgreSQL - Task DB]
  C --> G[Redis - Queue and Cache]
  C --> I[etcd - Service Registry & Leader Election]
  D --> G
  D --> I
  C --> H[Monitoring - Prometheus]
  D --> H
```

## 🚀 Features

* Dynamic task dispatching with Redis queue
* AI-based task prioritization (Python ML model)
* gRPC-based AI prediction integration
* Metrics tracking with Prometheus exporters
* Frontend task status view with Next.js📌 Features
* PostgreSQL storage and logging
* Leader election via etcd
* Prometheus metrics endpoint
* Delayed task queue via Redis ZSET
* RESTful APIs with CORS support
* Horizontally scalable worker model

## 🛠️ How to Run

This project provides a `Makefile` to simplify common tasks.

### 🚀 Run All Services (Backend + AI + Frontend)
```bash
make dev
```

This will:
- Start Docker services (Redis, PostgreSQL, etc.)
- Install Python dependencies
- Generate proto files
- Launch backend services (Go)
- Launch AI prediction service (Python)
- Launch frontend (Next.js)

### 🔄 Run Services Individually

```bash
make api         # Run Go API service
make scheduler   # Run Go scheduler service
make worker      # Run Go worker service
make ai          # Run Python AI prediction server
make client      # Run frontend (Next.js)
```


## 🚀 Performance Benchmark

Tested with [`wrk`](https://github.com/wg/wrk):

```bash
wrk -t12 -c400 -d30s -s test.lua http://localhost:8080/tasks
```

### ⚡️ Throughput

- **5660+ TPS** (tasks per second) sustained under pressure
- Equivalent to **340,000+ tasks per minute**
- End-to-end latency remains stable (~44ms average)


## 📎 License

MIT
