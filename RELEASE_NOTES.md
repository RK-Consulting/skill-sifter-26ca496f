# SkillSifter v0.5.6

## Business Dev Consolidation and Assignments Cleanup

SkillSifter v0.5.6 simplifies the recruiter-facing navigation by removing
two redundant modules. Assignments duplicated candidate-requirement
matching already handled by Daily Tasks. Business Dev has been folded
into Client — its two distinguishing fields, partner name and contact
person, are now simply part of the Client record, removing a separate
tab and form for data that belongs on the same entity.

## Highlights

**Business Dev merged into Client.**

- `partner_name` and `contact_person` are now optional fields on
  `Client`, added via migration `013_client_partner_and_contact_person.sql`.
- The standalone Business Dev tab, list view, and add-form are removed.
  All Business Dev data entry now happens through the existing
  Clients tab.

**Assignments removed.**

- The Assignments tab duplicated recruiter-requirement matching
  functionality already covered by Daily Tasks. Removed to reduce
  redundant UI surface as the tab count grows.

**Navbar overflow fix.**

- The nav-links section now scrolls independently of the logo and
  right-side action buttons (Add Candidate, Logout), which stay
  pinned and visible regardless of how many tabs are active.

## Known limitations

- The backend `BusinessDev` struct and its database table are now
  unused but not yet removed — left in place for this release and
  flagged for cleanup alongside future schema housekeeping.

## Testing & CI

No changes to the CI quality gate in this release. `gofmt`, `go build`,
`go vet`, `go test ./...` and the frontend lint/test/build pipeline
continue to gate every change.

## Next

Production is still running commit `643e33c`, predating both v0.5.5
and this release — redeploy to catch production up remains outstanding.

---

See [CHANGELOG.md](CHANGELOG.md) for the itemized change history.
