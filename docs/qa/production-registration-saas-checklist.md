# SkillSifter — Production Registration / SaaS QA Checklist

**Status:** v1.0.0 production validation  
**Scope:** Registration, OTP, tenant provisioning, permanent registration identity, login recovery, paid checkout  
**Last updated:** 2026-10-07

## A. CI / Release Gate
- [ ] Backend CI — PASS / FAIL
- [ ] Lint — PASS / FAIL
- [ ] `main` contains intended release commit
- [ ] `infra/scripts/deploy.sh` completed successfully
- [ ] `skillsifter` systemd service is ACTIVE
- [ ] API health-check returns HTTP 200
- [ ] Database migrations completed successfully
- [ ] Migrations 043 and 044 applied

## B. Fresh Trial Registration
Use a completely **new email address and new company**.
- [ ] Registration page loads
- [ ] Username entered
- [ ] New email entered
- [ ] Password entered
- [ ] Company name entered
- [ ] Trial/plan selected
- [ ] Registration request accepted
- [ ] Email OTP received
- [ ] OTP input is visible while typing
- [ ] Correct OTP accepted
- [ ] Registration verification succeeds

## C. Tenant Creation / Provisioning
- [ ] `platform_tenants` record created
- [ ] Tenant ID generated
- [ ] Tenant is created **before** tenant-bound user
- [ ] User created with correct `tenant_id`
- [ ] Trial subscription created
- [ ] Platform user account created
- [ ] Permanent registration registry record created
- [ ] Tenant database created
- [ ] Tenant migrations completed
- [ ] Tenant provisioning status = `READY`

## D. Login / Dashboard
- [ ] Newly registered admin can log in
- [ ] Login resolves the correct tenant
- [ ] Login does not require client-supplied tenant selection
- [ ] Dashboard loads
- [ ] Tenant data is accessible
- [ ] No provisioning error is shown
- [ ] Admin permissions work

## E. Duplicate Email — CRITICAL
Attempt registration again using the **same email**.
- [ ] Same email + different company → HTTP 409 / rejected
- [ ] Same email + different username → HTTP 409 / rejected
- [ ] Same email + different password → HTTP 409 / rejected
- [ ] Same email cannot create another tenant
- [ ] Different email can still register

The rejection must be based on the **permanent registration identity**, not merely on an existing tenant/user record.

## F. Original Registration Failure Regression
- [ ] Tenant is created before tenant-bound user
- [ ] No `users.tenant_id` foreign-key failure
- [ ] Provisioning failure does not destroy the registration identity
- [ ] Failed provisioning tenant remains recoverable
- [ ] Provisioning recovery authentication is possible
- [ ] `/admin/tenant/provision` is accessible to an authorized admin
- [ ] Re-running provisioning can make tenant `READY`
- [ ] Normal login works once tenant is `READY`

**Production safety:** Do not deliberately break the production database or provisioning infrastructure. Validate recovery through existing tests or controlled non-destructive testing.

## G. OTP Failure Cases
- [ ] Wrong OTP is rejected
- [ ] Expired OTP is rejected
- [ ] Correct OTP is accepted
- [ ] OTP input displays entered digits
- [ ] Successful verification creates permanent registry identity

## H. Paid Subscription
- [ ] Trial admin can open Account / Subscription
- [ ] Paid plan can be selected
- [ ] Checkout without phone verification is blocked
- [ ] Phone number can be entered
- [ ] Phone OTP can be sent
- [ ] Correct phone OTP accepted
- [ ] `phone_verified_at` is populated
- [ ] Checkout becomes available
- [ ] Razorpay checkout opens
- [ ] Successful provider confirmation/webhook activates subscription
- [ ] Account shows active paid subscription

## I. Tenant Isolation
- [ ] Tenant A cannot access Tenant B data
- [ ] Client cannot select another `tenant_id`
- [ ] Client cannot select another tenant database
- [ ] Admin privileges remain tenant-scoped
- [ ] RBAC is evaluated inside authenticated tenant context

## J. Trial Expiry / Data Deletion
Long-term lifecycle validation; not required to wait through the full period for every release.
- [ ] Trial expires
- [ ] Tenant becomes `EXPIRED`
- [ ] Data deletion period begins
- [ ] Tenant database is eventually deleted
- [ ] Tenant control-plane data is deleted
- [ ] Tenant-owned records are removed
- [ ] `platform_registration_registry` record remains
- [ ] Same email remains permanently registered
- [ ] Same email cannot register again

## K. Final Go-Live Decision
- [ ] CI GREEN
- [ ] Deployment GREEN
- [ ] Health-check GREEN
- [ ] Fresh registration GREEN
- [ ] OTP GREEN
- [ ] Tenant provisioning GREEN
- [ ] Login GREEN
- [ ] Dashboard GREEN
- [ ] Duplicate email rejection GREEN
- [ ] Paid checkout phone verification GREEN
- [ ] No tenant-isolation issue found

### Final result
- [ ] **PASS — READY**
- [ ] **FAIL — DO NOT RELEASE**

### Tester notes
____________________________________________________________

____________________________________________________________

____________________________________________________________

## Release acceptance path
```text
NEW EMAIL
   ↓
REGISTER
   ↓
EMAIL OTP
   ↓
VERIFY
   ↓
TENANT READY
   ↓
LOGIN
   ↓
DASHBOARD
   ↓
REGISTER SAME EMAIL AGAIN
   ↓
HTTP 409 / REJECTED
```

This proves both the registration/provisioning flow and the permanent registration-identity rule.