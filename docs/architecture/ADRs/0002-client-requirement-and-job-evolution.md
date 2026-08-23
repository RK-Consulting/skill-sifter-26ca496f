# ADR 0002: Client, Requirement, and Existing Job Evolution

**Status:** Proposed  
**Decider:** Product owner  
**Related issue:** V04-03

## Context

The current `jobs` resource does not model a recruitment-agency client requirement. V1 requires explicit client and requirement relationships.

## Decision required

Decide whether `jobs` evolves into `requirements`, whether a new `requirements` model is introduced with a controlled migration, and how legacy data and APIs remain compatible.
