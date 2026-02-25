---
name: dependency-audit
description: Structured workflow for auditing and updating project dependencies. Use when running security audits, updating packages, or triaging vulnerability reports.
---

# Dependency Audit Workflow

Use this skill when auditing dependencies for security vulnerabilities, planning updates, or triaging CVE reports.

## When to Use

- Scheduled monthly dependency review
- CI pipeline flags a vulnerability
- Planning a major version upgrade
- New CVE reported for a project dependency

## 1. Run Audit Commands

### Go
```bash
go mod verify          # Verify module checksums
govulncheck ./...      # Check for known vulnerabilities
go list -m -u all      # List available updates
```

### TypeScript/JavaScript
```bash
npm audit              # Check for vulnerabilities
npm outdated           # List available updates
npx npm-check-updates  # Show major updates available
```

### Python
```bash
pip audit              # Check for vulnerabilities (install: pip install pip-audit)
pip list --outdated    # List available updates
safety check           # Alternative vulnerability scanner
```

## 2. Triage Results

For each vulnerability or outdated package:

### Security Patches (Critical/High severity)
- [ ] Apply immediately — no waiting for release cycle
- [ ] Run full test suite after update
- [ ] Deploy to staging, verify, then production
- [ ] Document the CVE and fix in commit message

### Minor Version Updates
- [ ] Review changelog for breaking behavior changes (despite semver)
- [ ] Update and run tests
- [ ] Batch with other minor updates in a single PR
- [ ] Schedule for monthly update cycle

### Major Version Updates
- [ ] Create dedicated branch: `chore/upgrade-<package>-v<version>`
- [ ] Read migration guide thoroughly
- [ ] Identify breaking changes affecting your codebase
- [ ] Update code to match new API
- [ ] Run full test suite
- [ ] Test in isolation before merging

## 3. Decision Tree

```
Is there a known CVE?
  YES → Is it Critical or High severity?
    YES → Patch immediately, deploy today
    NO → Schedule for next sprint
  NO ↓

Is this a major version behind?
  YES → Create upgrade branch, plan migration
  NO ↓

Is this a minor/patch update?
  YES → Batch with other updates, test, PR
  NO → No action needed
```

## 4. Post-Audit Checklist

- [ ] All Critical/High vulnerabilities resolved
- [ ] Audit command produces clean output (or remaining items documented with justification)
- [ ] `go.sum` / `package-lock.json` / `poetry.lock` committed
- [ ] CI pipeline includes audit step that blocks on Critical vulnerabilities
