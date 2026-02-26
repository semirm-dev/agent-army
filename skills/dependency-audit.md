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

## 5. Transitive Dependency Vulnerabilities

When a vulnerability is in a transitive (indirect) dependency:

1. **Check if direct dependency has a fix:**
   - Go: `go list -m -json all | jq 'select(.Indirect)'` to find path
   - Node: `npm ls <vulnerable-package>` to find dependency chain
   - Python: `pipdeptree -r -p <vulnerable-package>` to find reverse deps

2. **If direct dep has updated:** Update the direct dependency
3. **If direct dep has NOT updated:**
   - Open an issue on the direct dependency's repo
   - Use resolution/override to force transitive version:
     - npm: `"overrides"` in package.json
     - Go: `replace` directive in go.mod
     - Python: pin the transitive dep directly in requirements

4. **Document the override** with a comment explaining why and a link to the upstream issue

## 6. SBOM (Software Bill of Materials)

Generate an SBOM for production deployments:

### Commands
```bash
# Go
go version -m <binary> > sbom.txt
# Or use: cyclonedx-gomod mod -output sbom.json

# Node
npx @cyclonedx/cyclonedx-npm --output-file sbom.json

# Python
pip-audit --format=cyclonedx-json > sbom.json
# Or: cyclonedx-py environment > sbom.json
```

### When to Generate
- Every production release
- Before security audits
- When onboarding to a new compliance framework

### What to Include
- Direct and transitive dependencies with versions
- License information
- Known vulnerability status

## 7. Supply Chain Security

- **Verify package provenance:** Use `npm audit signatures` (npm 9+), `go mod verify`
- **Lock files:** Always commit lock files (`package-lock.json`, `go.sum`, `poetry.lock`)
- **Pin versions:** Use exact versions in production, not ranges
- **Review new dependencies:** Before adding, check: maintenance activity, download counts, known vulnerabilities, license compatibility
- **Minimal dependencies:** Prefer stdlib over external packages for simple operations
- **Signed packages:** Where available, verify GPG signatures on releases
