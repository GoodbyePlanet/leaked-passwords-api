---
title: Storage strategy for pwned passwords
version: 1
date: 2025-09-25
authors: [Nemanja]
---

# ADR-001: Storage Strategy for Pwned Password Dataset

#### Context

We need to implement a pwned password check service. The dataset is large (~40GB), and we will have updates periodically (once a month).
The service must support fast lookups during authentication flows. We must decide how and where to store and query the dataset.
The goal is to not use directly [HaveIBeenPwnedAPI](https://haveibeenpwned.com/API/v3) to avoid hitting rate limits and
also to not have a dependency on the k-anonymity API.


### Options

### Option 1: SSD + Prefix Files + Bloom Filter (Single-node)

Store per-prefix files (SHA-1 first 5 hex chars → file) on local NVMe/Persistent SSD.
(There are 1,048,576 different hash prefixes between 00000 and FFFFF (16^5)).
Add a Bloom filter in memory to reduce disk lookups. Dataset refreshed monthly via scheduler.

#### Pros

- Simple to implement.
- Very fast local disk lookups.
- Low infra cost (just one VM with SSD).

#### Cons

- Not horizontally scalable (each VM must replicate full dataset).

### Option 2: SSD + BadgerDB (Embedded KV store)

- Store full dataset in BadgerDB (LSM-based embedded DB, optimized for SSD).
- Use SHA-1 hash as key.
- Refresh BadgerDB monthly by scheduled task that will download full pwned passwords hashes.

#### Pros

- Faster, more efficient lookups vs plain files.
- Simple local deployment (no external DB infra).
- Better organization than millions 16^5 of flat files.

#### Cons

- Still limited to a single-node model (data local to each VM).
- Rebuilding DB monthly requires downtime or careful swap strategy.

### Option 3: Centralized Real DB (Scaling-friendly)

- Import dataset into a shared DB (Postgres, DynamoDB, Cassandra, Elasticsearch, etc.).
- Application instances query the DB for lookups.
- Bloom filter is optional (in memory) to reduce DB load.

#### Pros

- Easy horizontal scaling — multiple API instances share one backend DB.
- Easier operations (no need to replicate dataset per instance).

#### Cons

- DB infra adds cost and complexity.
- Higher latency vs local SSD lookups.

### Option 4: S3 + Bloom (Hybrid, lightweight nodes)

- Store prefix files in S3/GCS bucket.
- Each instance keeps a Bloom filter locally to filter negative cases.
- On Bloom positive → fetch prefix file from S3. Cache locally for reuse.

#### Pros

- Lightweight nodes, no need to store full dataset locally.
- Centralized storage in S3 (easy updates).
- Elastic scaling (good for serverless / multi-instance).

#### Cons

- First lookup for a prefix incurs S3 latency.
- Requires caching layer for good performance.
- Slightly more complex than Option 1/2.

### Decision

For initial implementation, we will start with **Option 2** (SSD + BadgerDB (Embedded KV store)) for simplicity.
If scaling beyond a single node is required, we will evaluate Option 3 (Real DB) or Option 4 (S3 + Bloom).

### Status

Accepted
