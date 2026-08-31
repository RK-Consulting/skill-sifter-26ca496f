# ADR 0008: API and Domain Conventions

**Status:** Accepted  
**Decider:** Product owner  
**Related issue:** #24

## Context

SkillSifter currently exposes a mixture of canonical `/api/...` routes and compatibility routes registered at the root. The backend also uses a legacy response envelope (`success`, `message`, optional `data`) and tenant filtering based on the authenticated company context. V1 needs deterministic conventions that can be applied consistently without introducing an immediate public API break.

Issue #24 is architecture/documentation scope only. No existing public endpoint is removed or renamed by this decision.

## Decision

SkillSifter V1 adopts the following API and domain conventions.

### 1. API prefix and versioning

- The **canonical V1 API namespace is `/api/v1`**.
- New V1 endpoints must be introduced under `/api/v1/...`.
- Existing `/api/...` endpoints remain supported during the compatibility period and are not removed as part of this ADR.
- Existing root-level compatibility routes may remain temporarily where already deployed, but new endpoints must not be added there.
- A future major API version uses a new explicit version namespace rather than silently changing the contract of an existing version.

Example:

```text
GET /api/v1/candidates
GET /api/v1/jobs/{id}
POST /api/v1/interviews
```

### 2. Resource and naming conventions

- URLs use plural, lower-kebab-case resource names.
- URL path identifiers use `{id}` unless a more specific immutable identifier is required.
- HTTP methods express the operation: `GET`, `POST`, `PUT`, `PATCH`, and `DELETE`.
- JSON field names use **camelCase**, matching the existing public model convention.
- Go/domain model names use idiomatic Go `PascalCase` and `camelCase` field names.
- Database identifiers remain `snake_case`.
- Domain status values are explicit, documented strings; clients must not infer state from HTTP response wording.

### 3. Standard response envelope

The existing public response shape is retained for compatibility:

```json
{
  "success": true,
  "message": "...",
  "data": {}
}
```

Successful responses should use `data` for the resource or result set and a concise human-readable `message` where useful.

Errors use the same top-level envelope for the compatibility period:

```json
{
  "success": false,
  "message": "Resource not found"
}
```

Future V1 implementation work may add machine-readable error metadata, but must do so additively and must not make existing consumers dependent on a breaking replacement of `success`/`message`/`data`.

HTTP status codes remain authoritative for programmatic error classification.

Minimum conventions:

- `400` malformed or invalid request
- `401` unauthenticated
- `403` authenticated but not authorized
- `404` resource unavailable to the caller
- `409` resource/state conflict
- `422` semantically invalid input where the endpoint distinguishes it from malformed input
- `500` unexpected server failure

### 4. Filtering and pagination

Collection endpoints must use deterministic, bounded pagination once their result sets can grow materially.

Canonical query parameters:

```text
?page=1&limit=50
```

Rules:

- `page` is 1-based.
- `limit` has a server-defined maximum; clients cannot request an unbounded result set.
- Results have a deterministic ordering. The default ordering should be by a stable primary key or timestamp plus a stable tie-breaker.
- Filtering uses named resource fields, for example `status`, `location`, or other explicitly documented fields.
- Unknown filter fields are rejected rather than silently ignored when an endpoint defines a strict filter contract.
- Pagination and filtering are applied within the caller's authorized tenant scope.

When a response requires pagination metadata, the preferred shape is additive to the existing envelope:

```json
{
  "success": true,
  "message": "Candidates retrieved",
  "data": [],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 125
  }
}
```

### 5. Authorization

Authorization is enforced server-side and is not inferred from UI visibility or client-provided role values.

- Authentication establishes the user identity, tenant identity, and role.
- Role checks use the V1 RBAC policy defined by ADR 0005.
- `401` means no valid authenticated identity.
- `403` means the authenticated identity lacks the required permission.
- Resource ownership and tenant scope are checked independently of role checks.
- A client-supplied role, company name, tenant ID, or similar field must never elevate authorization.

### 6. Tenant scoping

ADR 0001 is authoritative for tenant identity and isolation. The API convention is therefore:

- Every tenant-owned resource is implicitly scoped to the authenticated tenant.
- The authenticated tenant ID is obtained from trusted authentication/request context.
- Client-supplied tenant identifiers or company names cannot override that context.
- Handlers, services, and repositories must preserve the tenant scope across the request boundary.
- Cross-tenant resource access must not reveal whether the resource exists.
- `company_name` is business/display data, not the long-term authorization key.

An endpoint must not require a client to send a tenant selector merely to access its own tenant's resources unless the endpoint is explicitly an administrative cross-tenant operation protected by higher-level authorization.

### 7. Compatibility policy

Compatibility is a deliberate V1 constraint.

- No public endpoint is removed as part of this ADR.
- Existing `/api/...` contracts remain supported while consumers migrate to `/api/v1/...`.
- Existing JSON field names and response envelope fields are preserved unless a separately approved compatibility decision is made.
- Breaking changes require a new API version or an explicit migration decision.
- Additive response fields are preferred over replacement fields.
- Deprecation must be documented before an existing public contract is retired.
- Internal implementation refactoring is allowed provided the externally observable contract remains compatible.

### 8. Domain boundary

API handlers translate HTTP concerns into application/domain operations. Domain logic must not depend on HTTP-specific details.

The preferred flow is:

```text
HTTP request
   -> authentication / authorization
   -> request validation
   -> application/service operation
   -> domain rules
   -> repository/data access
   -> response mapping
```

Handlers should remain thin. Business rules, state transitions, tenant enforcement, and persistence behaviour belong below the HTTP boundary.

## Existing implementation alignment

The current repository already has a shared JSON response helper using `success`, `message`, and optional `data`, so this ADR formalizes that compatibility contract rather than replacing it. fileciteturn99file0

The current routing exposes `/api` routes and some compatibility routes at the root. This ADR makes `/api/v1` the canonical namespace for new V1 work without requiring a public break now. The existing tenant architecture already establishes immutable tenant identity as the target security boundary and prohibits client-supplied tenant overrides. fileciteturn80file0 fileciteturn106file0

## Testing requirements

Subsequent implementation work must verify at minimum:

- canonical `/api/v1` routing;
- compatibility of existing `/api` routes;
- consistent JSON naming;
- standard success/error envelopes;
- correct HTTP status codes;
- bounded pagination and deterministic ordering;
- filtering cannot escape tenant scope;
- role enforcement returns `401`/`403` appropriately;
- known resource IDs cannot be used for cross-tenant access;
- client-supplied tenant or role fields cannot override trusted request context.

## Implementation boundary

This ADR defines conventions only. It does not authorize broad route migration, schema redesign, tenant migration, or public API removal. Those changes require separately scoped implementation issues.

## Consequences

### Positive

- New APIs have a predictable V1 contract.
- Existing consumers are protected during migration.
- Tenant isolation becomes an explicit API convention rather than an implicit handler responsibility.
- Pagination and filtering become deterministic and bounded.
- HTTP concerns remain separated from domain logic.

### Trade-offs

- `/api` compatibility must be maintained while consumers migrate to `/api/v1`.
- Some existing endpoints will temporarily coexist with the canonical V1 namespace.
- Future implementation work must gradually consolidate legacy route registration without breaking deployed clients.
