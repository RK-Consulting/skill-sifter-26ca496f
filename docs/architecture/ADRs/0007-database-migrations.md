# ADR 0007: Database Migration Strategy

**Status:** Proposed  
**Decider:** Product owner  
**Related issue:** V04-09

## Context

Current startup and deployment paths do not apply the same migration set. V1 needs deterministic fresh-install and upgrade behaviour.

## Decision required

Choose one migration runner and define ordering, tracking, idempotence, fresh-install bootstrap, failure handling, backup/rollback expectations, and test strategy.
