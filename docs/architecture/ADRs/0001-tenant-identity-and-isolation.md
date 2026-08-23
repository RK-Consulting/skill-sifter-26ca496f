# ADR 0001: Tenant Identity and Isolation

**Status:** Proposed  
**Decider:** Product owner  
**Related issue:** V04-02

## Context

The current shared schema uses mutable `company_name` values and handlers are individually responsible for tenant filters. V1 needs a reliable, testable tenant boundary.

## Decision required

Choose whether to retain `company_name` temporarily, migrate to an immutable tenant identifier, or use a staged compatibility approach. Define the authoritative enforcement point and safe migration path.

## Consequences to assess

Existing data migration, JWT claims, API compatibility, query conventions, cross-tenant tests, rollback, and tenant rename behaviour.
