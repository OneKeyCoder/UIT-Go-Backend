# Architecture Decision: Async Messaging Strategy

## Current State (As-Is)

### Services:

1. **broker-service** - API Gateway
2. **authentication-service** - JWT auth
3. **logger-service** - Centralized logging (MongoDB)
4. **listener-service** - RabbitMQ consumer → forwards to logger
5. **mail-service** - Email scaffold (unused)

### Communication Patterns:

```
Synchronous (HTTP):
- Client → Broker → Auth (immediate response needed)
- Client → Broker → Logger (direct HTTP)
- Auth → Logger (fire-and-forget logging)

Asynchronous (RabbitMQ):
- Broker → RabbitMQ → Listener → Logger (queued logging)
```

### Problem:

**TWO ways to log:**

1. Direct HTTP: `broker → logger-service`
2. Queue: `broker → RabbitMQ → listener → logger-service`

This is **redundant and confusing**.

---

## Decision: Keep Hybrid Architecture (Recommended)

### Strategy:

**Use the right tool for each job**

#### ✅ Synchronous HTTP/gRPC (Immediate Response)

Use for:

-   Authentication (need immediate token)
-   Ride requests (rider needs instant confirmation)
-   Driver location updates (real-time)
-   Payment processing (need immediate success/failure)

```
Client → Broker → Service → Immediate Response
```

#### ✅ Asynchronous RabbitMQ (Fire-and-Forget)

Use for:

-   Logging (don't wait for DB write)
-   Notifications (send email/SMS later)
-   Analytics events (process in background)
-   Audit trails (eventual consistency OK)

```
Service → RabbitMQ → Listener → Background Processing
```

---

## Updated Architecture for Uber App

```
┌─────────────────────────────────────────────────────────────┐
│                         CLIENT                               │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│              BROKER (API Gateway)                            │
│  - JWT Validation                                            │
│  - Rate Limiting                                             │
│  - Request Routing                                           │
└──┬────────┬─────────┬─────────┬──────────┬─────────────┬────┘
   │        │         │         │          │             │
   ▼        ▼         ▼         ▼          ▼             ▼
┌─────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌───────────┐ ┌──────────┐
│Auth │ │Rider │ │Driver│ │ Ride │ │ Location  │ │ Payment  │
│     │ │      │ │      │ │      │ │ (Redis)   │ │          │
└──┬──┘ └───┬──┘ └───┬──┘ └───┬──┘ └─────┬─────┘ └────┬─────┘
   │        │        │        │          │            │
   │        └────────┴────────┴──────────┴────────────┘
   │                      │
   │                      ▼
   │              ┌───────────────┐
   │              │   RabbitMQ    │
   │              └───────┬───────┘
   │                      │
   ▼                      ▼
┌──────────────┐   ┌──────────────┐
│Logger Service│   │Listener      │
│(MongoDB)     │   │- Email Queue │
└──────────────┘   │- SMS Queue   │
                   │- Audit Queue │
                   └──────────────┘
```

---

## Implementation Plan

### Keep:

-   ✅ **Listener Service** - For async work
-   ✅ **RabbitMQ** - For message queuing
-   ✅ **Logger Service** - Centralized logs

### Remove:

-   ❌ Direct HTTP logging from services to logger
-   ❌ Mail service from docker-compose (keep scaffold)

### Add:

-   ✅ Notification service (email/SMS via RabbitMQ)
-   ✅ Audit service (compliance logs via RabbitMQ)

### Refactor:

-   🔧 **All services** → Log via RabbitMQ only (not HTTP)
-   🔧 **Listener** → Handle multiple queue types:
    -   `log.*` → Logger Service
    -   `notification.*` → Notification Service (email/SMS)
    -   `audit.*` → Audit Service

---

## Queue Topics

```
log.INFO       → Logger (informational)
log.WARNING    → Logger (warnings)
log.ERROR      → Logger (errors)

notification.email  → Notification Service
notification.sms    → Notification Service
notification.push   → Notification Service

audit.ride_created  → Audit Service
audit.payment_made  → Audit Service
audit.user_action   → Audit Service
```

---

## Benefits

### Performance

-   ⚡ Services don't wait for logging to complete
-   ⚡ Non-blocking notifications
-   ⚡ Handles traffic spikes (queue buffers)

### Reliability

-   🛡️ Messages survive service crashes (queue persistence)
-   🛡️ Retry failed operations automatically
-   🛡️ Prevents cascade failures

### Scalability

-   📈 Run multiple listener instances (horizontal scaling)
-   📈 Process 10K+ messages/second
-   📈 Add new consumers without touching producers

### Development

-   🔧 Easy to test (mock queue)
-   🔧 Add new event types without changing services
-   🔧 Demonstrates knowledge of event-driven architecture

---

## Trade-offs

### Pros:

-   ✅ Production-ready architecture
-   ✅ Shows understanding of async patterns
-   ✅ Impressive for school project

### Cons:

-   ❌ More complex than direct HTTP
-   ❌ Requires RabbitMQ infrastructure
-   ❌ Need to monitor queue depths

---

## Verdict

**KEEP the hybrid architecture** because:

1. **Educational Value** - Shows you understand when to use sync vs async
2. **Real-world Pattern** - Uber actually uses message queues (Kafka)
3. **Scalability** - Ready for high traffic
4. **Impressive** - Professors will appreciate the thoughtfulness

But **document it clearly** so team understands the "why" behind each choice.

---

## Next Steps

1. ✅ Remove direct HTTP logging calls
2. ✅ Expand listener to handle multiple topics
3. ✅ Add notification service (email/SMS queue consumer)
4. ✅ Document queue contract in code comments
5. ✅ Add queue monitoring dashboard (optional)

---

_Decision made: October 4, 2025_  
_Rationale: Over-engineered but demonstrates knowledge (user preference)_
