# Jatify - Job Application Tracker

### Main Feature
- Job Application Tracking
- AI Resume Analyzer with Open Router's Model
- Job Queue System for Resume Analysis with DLQ fallback implementation
- Concurrent Job Processing for Resume Analysis

### Repository's Architecture
Implemented using clean architecture principles, emphasizing on separation of concerns using different types of layers. In general, there are **four** layers
| Layer Name | Responsibilities |
| -----------|-------------------|
| Handler Layer | HTTP Request Validation, Request Processing  |
| Service Layer | Main Business Process |
| Repository Layer | Interacting with database or anything data related |
| Entity Layer | Struct, Data definition across other layers |

### Visualization of The Architecture
<img width="406" height="219" alt="image" src="https://github.com/user-attachments/assets/4195ba23-7941-4b00-a6d0-32ed927ff944" />


## Job Queue Design
Note that the term "Job" here is not the actual Job as in the main Job Application Feature. In this section, "Job" term is defined as an object that needed to be processed for the 
AI Resume analyzer. 
# Simplified Enqueue & Requeue Flow
The Job Queue sytem is designed with concurrency, utilizing the go routine feature from Go. Right now, the concurrent workers are hardcoded into three workers only.
In a nutshell, this three workers will "race" to find the next "Job" available. The next job available are defined as follows

- The latest pending job in the database

Every /enqueue calls will populate a channel. A channel input would wake one of N (3 for now) concurrent workers to do the process. The process will claim the next Job i.e the latest pending jobs in the database

## Enqueue

```mermaid
flowchart LR
    A[Client] --> B[API: analyze resume]
    B --> C[(jobs table\nstatus=pending)]
    C -->|ring the doorbell| D[jobCh]
    C -. backoff timer expires .-> E
    D -->|worker wakes up| E[ClaimNext\noldest pending job]
    E --> F[(jobs table\nstatus=done)]
```

## Requeue (from DLQ)

```mermaid
flowchart LR
    A[(jobs table\nstatus=failed)] --> B[(DLQ)]
    B --> C[Client requeue]
    C --> D[(jobs table\nstatus=pending)]
    D -. backoff timer expires .-> E[ClaimNext\noldest pending job]
    E --> F[(jobs table\nstatus=done)]
```

---

## How the Worker Wakeup Works

```mermaid
flowchart TD
    A[Worker idle\nblocked on select] --> B{which fires first?}
    B -->|channel receives a signal| C[Fast path\nwake up immediately]
    B -->|backoff timer expires| D[Slow path\nwake up on schedule]
    C --> E[ClaimNext\noldest pending job from DB]
    D --> E
    E -->|job found| F[Process job\nstatus = processing]
    E -->|nothing found| G[backoff doubles\nup to 30s cap]
    F --> H[MarkDone or MarkFailed]
    H --> A
    G --> A
```

---

## Glossary

| Term | What it does |
|---|---|
| **jobCh** | Buffered channel (cap 30). Acts as a doorbell — populated on enqueue, discarded by worker. Purely a wakeup signal, not a data queue. |
| **Fast path** | Worker woken by a channel signal. Goes straight to `ClaimNext()`. |
| **Slow path** | Worker woken by backoff timer expiring. Also goes to `ClaimNext()`. Safety net for missed signals (channel full, requeue, crash recovery). |
| **ClaimNext** | Atomically grabs the oldest `pending` job (`ORDER BY created_at`), flips it to `processing`. Uses `FOR UPDATE SKIP LOCKED` so concurrent workers never claim the same job. |
| **Backoff** | Each worker's poll interval. Starts at 1s, doubles on empty poll, caps at 30s. Resets to 1s when a job is found. Prevents idle workers from hammering the DB. |
| **DLQ** | Dead Letter Queue. Permanently failed jobs land here. Requeue flips the original job back to `pending` — no channel signal sent, worker discovers it via slow path. |
| **FIFO** | Enforced by `ClaimNext` alone. Channel ID is always discarded — both paths go through DB, so queue order is never bypassed. |


