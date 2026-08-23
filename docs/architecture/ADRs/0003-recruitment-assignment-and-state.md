# ADR 0003: Recruitment Assignment and Transaction State

**Status:** Proposed  
**Decider:** Product owner  
**Related issues:** V04-04, V04-06

## Context

V1 must connect a candidate and a client requirement through a durable transaction. Candidate state and transaction state are separate concerns.

## Decision required

Define assignment ownership, cardinality, lifecycle states/transitions, snapshots, audit events, and the downstream ownership of submissions, interviews, offers, joining, and commercial records.
