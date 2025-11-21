# GCP Advanced Architectural Patterns

## BigQuery Architecture for Analytics

**Pattern Overview**: Petabyte-scale SQL analytics with automatic scaling.

**Components**:
```
┌──────────────────────────────────┐
│   Data Sources (Streaming)       │
│   Cloud Pub/Sub → Dataflow       │
└──────────────┬───────────────────┘
               │
        ┌──────▼──────┐
        │  BigQuery   │
        │  (Partitioned)
        └──────┬──────┘
               │
   ┌───────────┴────────────┐
   │                        │
┌──▼──┐              ┌──────▼─┐
│BI   │              │ GCS    │
│Tools│              │Export  │
└─────┘              └────────┘
```

**Cost Optimization**:
- Partition by date (only scan needed days)
- Cluster by frequently filtered columns
- Use BI Engine for cached results
- Approximate queries for top-k queries

---

## GKE (Google Kubernetes Engine) Multi-Region

**Architecture Pattern**:
```
┌─────────────────────────────────────┐
│   GKE Autopilot Multi-Region        │
├─────────────────────────────────────┤
│                                     │
│  us-central1 ◄────► eu-west1       │
│  (Primary)          (Secondary)     │
│                                     │
│  - Config Connector                 │
│  - Workload Identity (no keys!)     │
│  - Multi-cluster Ingress            │
│  - Anthos Service Mesh              │
│                                     │
└─────────────────────────────────────┘
```

**Benefits**:
- Automatic node provisioning
- Integrated identity (no service account keys)
- Native GCP service integration

---

## Firestore Real-Time Database Design

**Data Structure**:
```
/users/{userId}
  - name, email
  - /subscriptions/{subId}
    - status, created_at

/orders/{orderId}
  - user_id (reference)
  - items [array]
  - total, status
```

**Query Patterns**:
- Single document read: 1ms
- Collection query: 10-50ms
- Real-time listener: 100-200ms
- Transaction: 500-2000ms

---

## Vertex AI ML Pipelines

**End-to-End ML Workflow**:
```
┌─────────────┐
│ Data        │
│ Preprocessing
└──────┬──────┘
       │
   ┌───▼──────┐
   │ Feature  │
   │ Store    │
   └───┬──────┘
       │
   ┌───▼───────────┐
   │ Vertex AI     │
   │ Training      │
   └───┬───────────┘
       │
   ┌───▼──────┐
   │ Model    │
   │ Eval     │
   └───┬──────┘
       │
   ┌───▼──────┐
   │ Deploy   │
   │ Endpoint │
   └──────────┘
```

---

## Cloud Run Auto-Scaling

**Request Flow**:
1. Request arrives
2. Cloud Run auto-scales to 100 instances
3. Process completes
4. Scales down after 15 minutes idle

**Cost-Effective for**:
- Web APIs
- Event handlers
- Scheduled jobs
- Webhooks

---

**Learn More**: See optimization.md for specific cost reduction strategies.
