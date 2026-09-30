# Phase 9 — SaaS Subscription, Bridge Tenancy, and Login Context

**Status:** Architecture implemented and frozen for v1.0.0  
**Related issue:** #79  
**ADR:** ADR 0014  
**Current implementation status:** v1.0.0 production release  
**Authorization contract:** See [V1 Privilege Matrix](authorization/v1-privilege-matrix.md)

## Purpose

Phase 9 introduces the SaaS account layer for SkillSifter without changing the recruitment lifecycle.

Recruitment billing remains a SkillSifter operational record. Formal invoicing, GST/tax processing, accounting, banking, and client payment collection remain outside SkillSifter.

SkillSifter's own SaaS subscription is separate: a subscribing recruitment company gets a SkillSifter account, a subscription, an isolated tenant database, and tenant-scoped RBAC.

## Target architecture

~~~text
                         INTERNET
                            │
                     SkillSifter API
                            │
                    ┌───────┴────────┐
                    │                │
              CONTROL PLANE      AUTHENTICATION
                    │                │
          ┌─────────┼─────────┐      │
          │         │         │      │
        Tenant   Subscription Plan  User/RBAC
          │         │         │      │
          └─────────┴─────────┴──────┘
                            │
                    Tenant Resolver
                            │
                     Tenant Database
                            │
                  PostgreSQL shared cluster
                            │
              ┌─────────────┴─────────────┐
              │                           │
          Tenant A DB                 Tenant B DB
~~~

## Control plane

The control plane stores platform/account data:

- tenant ID
- company name
- company email
- company phone
- company address
- subscription plan
- subscription status
- subscription dates
- provisioning status
- tenant database identifier/routing metadata
- payment provider name
- opaque provider references

It does not store card numbers, CVV, UPI credentials, bank credentials, or other payment-instrument data.

## Tenant database

Each subscribed company gets one dedicated PostgreSQL database on the shared PostgreSQL infrastructure.

That database contains the company's SkillSifter operational data. Existing tenant-owned domains continue to use the same domain models; the physical database boundary replaces the current shared-schema tenant isolation model over the staged migration.

Subscription renewal reuses the existing tenant database.

## Subscription activation

~~~text
Register account
    ↓
Create control-plane tenant
    ↓
Select plan
    ↓
External payment-provider checkout
    ↓
Provider webhook/confirmation
    ↓
Activate subscription
    ↓
Provision tenant database
    ↓
Run tenant migrations
    ↓
Create initial admin
    ↓
Tenant ACTIVE
~~~

Provisioning must be idempotent.

## Login

Login establishes platform context before the user enters the recruitment application:

~~~text
Credentials
   ↓
Authenticate user
   ↓
Resolve trusted tenant
   ↓
Read subscription state
   ↓
Reject inactive tenant
   ↓
Resolve tenant database
   ↓
Resolve role/RBAC
   ↓
Create authenticated request context
   ↓
Tenant application
~~~

The client cannot choose the tenant database, tenant ID, company, or role.

## RBAC

The existing V1 roles remain:

- admin
- manager
- recruiter
- team_leader

RBAC is evaluated only after authentication establishes the tenant.

Tenant isolation and RBAC are separate checks:

~~~text
Authenticated identity
        │
        ├── tenant → which database/data
        │
        └── role   → which operations
~~~

An admin is an administrator of that tenant, not of all SkillSifter tenants.

## Payment-provider abstraction

The subscription domain remains payment-provider independent.

~~~text
Subscription Service
       │
Payment Provider Interface
       │
  ┌────┼────┐
  │    │    │
Juspay Razorpay Stripe
~~~

Providers return normalized subscription/payment events to SkillSifter. Provider-specific financial data remains with the provider.

## Implementation sequence

### 9A — Control plane

- tenant account model
- subscription/plan model
- tenant database metadata
- provisioning status

### 9B — Tenant database provisioning

- database creation abstraction
- deterministic tenant database naming/identity
- tenant migration execution
- idempotent provisioning

### 9C — Tenant database routing

- authenticated tenant → database resolution
- request-scoped database context
- removal of application dependence on one global tenant data connection

### 9D — Login + subscription + RBAC

- authenticate
- resolve tenant
- verify subscription access
- resolve role
- establish tenant DB context
- apply RBAC

### 9E — Payment provider adapter

- provider-neutral interface
- first provider adapter
- webhook normalization
- activation/renewal/failure/cancellation events

### 9F — Frontend account/subscription

- company account
- subscription plan
- subscription status
- payment-provider checkout handoff
- account access state

## Important implementation constraint

Do not refactor the application into microservices as part of Phase 9.

The first implementation remains a modular Go application. The bridge tenancy boundary is the architectural separation we need now; service decomposition can be introduced later only if justified by scale or operational requirements.

## Security acceptance criteria

- No client-supplied tenant value can select a tenant database.
- No user can authenticate into another company's tenant.
- Subscription state is checked before normal tenant application access.
- RBAC is evaluated within the authenticated tenant.
- Admin privileges never cross tenants.
- Tenant A database connections cannot be used for Tenant B requests.
- Payment instruments remain outside SkillSifter.
- Existing recruitment lifecycle tests remain green.
- Provisioning is idempotent.
- Tenant migration failures do not activate the tenant.

## v1.0.0 implementation status

The control-plane, subscription, tenant provisioning, tenant routing, authentication, RBAC and account/subscription UI architecture described here are implemented in the v1.0.0 production release. No microservice decomposition is introduced by this architecture.
