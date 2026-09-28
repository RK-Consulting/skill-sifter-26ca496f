# SkillSifter V1 — Privilege Matrix and Authorization Implementation

**Status:** Approved authorization contract; implementation partially enforced  
**Roles:** Admin, Manager, Team Leader, Recruiter  
**Related ADR:** ADR 0005  
**Tenant boundary:** ADR 0001 / ADR 0014 / ADR 0015

## 1. Authorization Principles

SkillSifter has two independent security decisions:

1. **Tenant authorization** — which tenant's data may be accessed.
2. **Role authorization** — what the authenticated user may do inside that tenant.

```
Request
  ↓
Authentication
  ↓
Trusted Tenant Identity
  ↓
Trusted Role
  ↓
Domain Action Authorization
  ↓
Tenant-scoped Operation
```

A client-supplied tenant ID, role, user ID or company value must never override authenticated context.

There is no V1 custom permission builder.

## 2. Fixed Roles

### Admin

Full tenant administration and operational access.

Only Admin can:

- create tenant users
- change tenant user roles
- delete tenant users
- manage tenant-level configuration

The initial tenant Admin cannot be promoted from another role.

### Manager

Operational management across the tenant.

Manager may manage operational recruitment data according to the matrix, but cannot administer tenant users.

### Team Leader

Recruitment/team operations within the defined team scope.

Team Leader must not silently become an unrestricted tenant administrator.

### Recruiter

Day-to-day assigned recruitment operations.

Recruiter cannot administer users, roles, or tenant configuration.

## 3. V1 Domain Action Matrix

| Domain | Admin | Manager | Team Leader | Recruiter |
|---|---|---|---|---|
| Users / roles | Full | Read | None | None |
| Clients | Full | Create/update | Read/use | Read/use assigned |
| Requirements | Full | Full | Create/use within team scope | Create/use assigned |
| Candidates | Full | Full | Full within team scope | Create/update/use assigned |
| Recruitment assignments | Full | Full | Full within team scope | Create/update assigned |
| Screening | Full | Full | Full | Full assigned |
| Submissions | Full | Full | Full | Full assigned |
| Feedback | Full | Full | Full | Full assigned |
| Interviews | Full | Full | Full | Full assigned |
| Selection | Full | Full | Full within team scope | Operational/assigned |
| Offers | Full | Full | Full within team scope | Operational follow-up |
| Joining | Full | Full | Full within team scope | Operational follow-up |
| Billing | Full | Full | Read/operational as authorized | No administrative billing access |
| Business Development | Full | Full | Read/use as authorized | Limited/use assigned |
| Reports | Full | Full | Team/operational | Operational/assigned |
| Resume / AI operations | Full | Full | Operational/team scope | Assigned recruitment scope |
| Daily Tasks | Full | Full | Team scope | Assigned scope |
| Tenant configuration | Full | Limited | None | None |

**Important:** "Full" never means cross-tenant access.

## 4. Current Enforcement Status

### Already enforced

- Admin-only user-management API
- fixed roles during public registration
- first user forced to Admin
- additional users limited to Manager / Recruiter / Team Leader
- subscription seat limit on additional users
- candidate create/update/delete role checks
- interview create/update role checks
- recruitment lifecycle mutation role checks on the currently wired V1 routes
- trusted role resolution from platform access
- subscription access checked before protected requests

### Still to enforce systematically

- complete endpoint-by-endpoint matrix
- team scope for Team Leader
- assigned scope for Recruiter
- Billing read/create separation
- Business Development
- Daily Tasks
- Reports
- Resume/AI operations
- Client and Requirement read/write distinctions
- Selection/Offer/Joining action distinctions
- tenant configuration
- cross-module negative authorization tests

## 5. Implementation Convention

Do not solve authorization by adding random role checks to every handler.

Use a consistent domain-action convention.

Conceptually:

```
domain action
    ↓
allowed roles
    ↓
tenant scope
    ↓
handler/service operation
```

The implementation may continue using the existing role middleware while migrating toward a consistent authorization convention.

Do not introduce:

- arbitrary permission JSON editors
- custom role creation
- policy languages
- enterprise authorization engines

## 6. Required Negative Tests

For every protected domain action:

1. unauthenticated request → rejected
2. wrong role → rejected
3. wrong tenant → rejected
4. client-supplied role → ignored/rejected
5. client-supplied tenant → ignored/rejected
6. Admin from Tenant A → cannot access Tenant B
7. Recruiter → cannot administer users
8. Team Leader → cannot become unrestricted tenant-wide admin
9. Manager → cannot create/promote/delete users
10. subscription inactive → protected tenant operation rejected

## 7. Completion Criteria

The privilege-matrix implementation is complete only when:

- every protected endpoint maps to an approved domain action
- every domain action has an explicit role policy
- tenant scope is enforced independently
- negative authorization tests exist
- frontend navigation/actions respect the same policy
- backend remains authoritative even if the frontend is bypassed
- no hidden role escalation path remains
