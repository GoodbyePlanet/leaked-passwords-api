# Badger DB Notes

## 1. What is Badger DB?

**BadgerDB** is an **embedded, high-performance key-value database** written in Go.
It's powerful, transactional, and concurrency-safe key-value store.
It is designed as a **pure Go alternative to LevelDB and RocksDB**, optimized for SSDs and concurrent workloads.
Badger provides **low-latency reads**, **fast writes**, and **crash safety**, making it ideal for applications that need
to persist large sets of key-value pairs without running a separate database service.

---

## 2. Writing Data with `txn.Set()` — Atomic Updates and MVCC

### `txn.Set()`

```go
txn.Set([]byte("key"), []byte("value"))
```

Badger adds or updates that key-value pair **inside a write transaction** (`db.Update`).

This operation is:

* **Buffered in memory** until the transaction commits.
* **Atomic**, meaning that either *all* writes in the transaction succeed, or *none* do.

### Atomic Updates

An atomic update ensures **data consistency**:

* If your application crashes mid-update, no partial or corrupted data remains.
* Only fully committed transactions appear to other readers.

### MVCC (Multi-Version Concurrency Control)

Badger uses MVCC to let **multiple readers** and **one writer** operate safely at the same time:

* Each read transaction (`db.View`) sees a **consistent snapshot** of the database at the time it started.
* Write transactions (`db.Update`) create new *versions* of changed keys.
* Old versions remain available for readers until the read transaction completes.
* Once no readers depend on old versions, Badger safely reclaims that space.

This allows concurrent downloads (writes) and API reads to coexist without locking or corruption.

---

## 3. File Extensions

Badger stores data on disk in two main file types:

| Extension | Name                    | Description                                                                                                                                                           |
| --------- | ----------------------- |-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `.sst`    | **Sorted String Table** | Immutable files that hold sorted key-value pairs. They are like an index (a phone book) that helps you quickly find where each value is.                              |
| `.vlog`   | **Value Log**           | File that hold the actual values — especially large ones that don’t fit nicely into `.sst`. The `.sst` files hold metadata and pointers to values inside `.vlog` files. |

This structure keeps reads fast (sorted lookups in `.sst`) and writes efficient (append-only `.vlog`).

---

## 4. Special Files

| File            | Purpose                                                                                                                                                                               |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **DISCARD**     | Tracks obsolete or deleted data to help Badger reclaim space during compaction.                                                                                                       |
| **KEYREGISTRY** | Holds metadata related to encryption keys if encryption is enabled. Even without encryption, it may exist as a placeholder.                                                           |
| **MANIFEST**    | Core metadata file describing the current state of the database — e.g., which `.sst` files exist and how they’re organized. It’s critical for crash recovery and startup consistency. |

---

## 5. What Happens When a New Download Starts (On app startup OR scheduler runs)

* Each `txn.Set()` with the same key **overwrites** the previous value.
* The update is atomic, ensuring no partial data is exposed.
* Readers (`db.View`) continue to see consistent data during writes — they either get the **old** or **new** value, never a mix.
