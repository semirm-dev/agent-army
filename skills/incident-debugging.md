---
name: incident-debugging
description: "Systematic debugging workflow — reproduce, isolate, hypothesize, verify, fix, prevent. Includes log analysis, root cause table, bisect strategy, and post-incident review."
scope: universal
uses_rules:
  - observability
  - cross-cutting
---

# Incident Debugging Skill

## When to Use

Invoke this skill when:
- Investigating a bug reported by users or QA
- Debugging a production incident (elevated error rate, degraded latency, outage)
- Analyzing error logs or alerting output
- Diagnosing intermittent or non-deterministic failures
- Conducting a post-incident review

> See `rules/observability.md` for structured logging field requirements and log level definitions.

## Debugging Workflow

Follow these six steps in order. Do not skip steps.

```
1. Reproduce --> 2. Isolate --> 3. Hypothesize --> 4. Verify --> 5. Fix --> 6. Prevent
     ^                                                |
     |                                                |
     +------------ Hypothesis disproved? -------------+
```

### Step 1: Reproduce

Establish a reliable reproduction before changing any code.

- **Gather context:** Error messages, stack traces, affected users/requests, timeline of occurrence, recent deploys.
- **Reproduce locally** with the same input, configuration, and data state. Use the exact request payload if available.
- **If you cannot reproduce locally:** Check environment differences (config, feature flags, data volume, concurrency level). Try staging. Check if the issue is load-dependent or timing-dependent.
- **Document reproduction steps** as a numbered list. These become the basis for your regression test.

### Step 2: Isolate

Narrow the scope systematically. Use binary search, not intuition.

```
Where does it fail?
  |
  +--> Frontend or Backend?
         |
         +--> Which service?
                |
                +--> Which endpoint / handler?
                       |
                       +--> Which function / line?
```

Techniques:
- **Comment out / disable** sections of code to isolate the failing path
- **Add boundary logging** at entry/exit of suspect functions (remove after debugging)
- **Use git bisect** for regressions (see Bisect Strategy below)
- **Check recent changes:** Review commits and deploys since the last known-good state
- **Divide external dependencies:** Swap real services for mocks to determine if the issue is internal or external

### Step 3: Hypothesize

Form 2-3 hypotheses based on evidence collected so far. Never guess without data.

For each hypothesis, document:
1. **What you think is happening** (one sentence)
2. **What evidence supports it** (log lines, metrics, stack traces)
3. **What evidence would confirm or disprove it** (specific test, log output, metric change)

Rank hypotheses by likelihood. Investigate the most likely first.

**Classify the error type:**
- Domain Error? Check validation logic, business rules, data integrity.
- Infrastructure Error? Check connections, timeouts, resource limits, downstream health.
- System Error? Check for nil/null dereferences, unhandled edge cases, concurrency bugs.

### Step 4: Verify

Test one hypothesis at a time. Changing multiple things simultaneously makes it impossible to identify the cause.

- **Add temporary logging** at the suspected failure point. Include relevant state (variable values, request context). Remove after debugging.
- **Use a debugger** for complex state issues (breakpoints, watch expressions, step-through).
- **Check metrics and traces** for the affected time window. Correlate by `request_id` or `trace_id` (see Log Analysis Patterns below).
- **If the hypothesis is disproved,** return to Step 3 with the new evidence. Update your hypothesis list.

### Step 5: Fix

Apply the minimal fix that addresses the root cause, not the symptom.

1. **Write the regression test first.** The test must fail before the fix and pass after. This proves the fix is correct and prevents recurrence.
2. **Make the smallest change possible.** Resist the urge to refactor adjacent code during an incident fix.
3. **Verify the fix does not introduce new failures.** Run the full test suite. Check for side effects in related functionality.
4. **Document the root cause** in the commit message or PR description. Future readers need to understand why this change was made.

### Step 6: Prevent

A bug that can recur is not fixed.

- **Add monitoring or alerting** for the failure mode so it is detected before users report it.
- **Update runbooks** if this is a production incident. Include symptoms, investigation steps, and resolution.
- **Check for similar bugs** in adjacent code. The same mistake pattern often exists in multiple places.
- **Consider systemic fixes:** Could a linter rule, type constraint, or architectural change prevent this class of bug entirely?

## "Can You Reproduce?" Decision Tree

```
Can you reproduce the bug?
  +-- YES --> Capture steps → reproduce locally → isolate (Step 2) → fix (Step 5)
  +-- NO  --> Do you have logs or traces?
                +-- YES --> Log analysis → reconstruct timeline → hypothesize → targeted repro
                +-- NO  --> Add logging at suspect boundaries + metrics + alerts → wait for
                            recurrence with instrumentation → capture data → restart from top
```

## Log Analysis Patterns

- **Find the first error:** Cascading failures produce many log entries. The first error in the timeline is usually the root cause. Filter by timestamp range, sort chronologically, look for ERROR or WARN level.
- **Correlate by request/trace ID:** Use `request_id` or `trace_id` to follow a single request across services.
- **Reconstruct timeline:** Identify the time window, collect logs from all services, sort by timestamp, find the transition from healthy to unhealthy.
- **Pattern match:** Look for repeated errors with increasing frequency (cascading failure), gaps in expected logs (crash, deadlock), correlation with external events (deploys, cron jobs), and resource metrics at the failure time.

## Common Root Causes by Symptom

| Symptom | Common Root Causes | Where to Look |
|---------|-------------------|---------------|
| **500 errors** | Unhandled exception, nil/null pointer dereference, DB connection pool exhausted, panic/unhandled promise rejection | Stack traces, error logs, connection pool metrics |
| **Timeouts** | Slow query (missing index), connection pool exhaustion, downstream service degraded, lock contention, large payload | Query plans (`EXPLAIN ANALYZE`), latency histograms, dependency health checks |
| **Memory leak** | Unclosed connections/file handles, unbounded cache growth, event listener accumulation, large object retention in closures | Memory profiler, heap dumps, connection pool metrics over time |
| **Intermittent failures** | Race condition, resource contention under load, network flakiness, GC pauses, connection pool size too small | Logs correlated with load metrics, `-race` flag (Go), thread dumps |
| **Data inconsistency** | Missing transaction boundary, concurrent writes without locking, cache serving stale data, replication lag | Transaction logs, cache TTL settings, replication metrics |
| **Slow degradation** | Memory leak, connection leak, log volume growth filling disk, unbounded queue growth | Trending metrics (memory, connections, disk) over hours/days |

## Bisect Strategy

Use `git bisect` when you know the bug was introduced between two known states (a working commit and a broken commit).

### When to Use

- The bug is a regression (it worked before, now it does not)
- You can write a test or script that reliably detects the bug
- The commit history between good and bad states is non-trivial (10+ commits)

### Running Bisect

1. `git bisect start` → `git bisect bad` (current) → `git bisect good <known-good-sha>`
2. Git checks out a middle commit. Test it. Mark `git bisect good` or `git bisect bad`.
3. Repeat until git identifies the first bad commit. Run `git bisect reset` when done.

**Automated:** Write a script that exits 0 for good and non-zero for bad, then run `git bisect run ./bisect-test.sh` to let git test each step automatically.

## Post-Fix Verification

After merging the fix:

1. **Verify the regression test passes in CI.** Do not rely on local runs alone.
2. **Deploy to staging.** Reproduce the original failure scenario. Confirm it no longer occurs.
3. **Deploy to production.** Monitor the specific metrics and log patterns from your investigation.
4. **Monitor for 24 hours.** Watch error rates, latency percentiles, and the specific metric you added in Step 6 (Prevent).
5. **Close the loop:**
   - Update the incident ticket with root cause, fix, and prevention measures.
   - Update the runbook if this is a new failure mode.
   - Share findings with the team (brief post-mortem for significant incidents).

## Post-Incident Review Template

For significant incidents, document:

1. **Summary:** What happened, when, and who was affected.
2. **Timeline:** Chronological sequence of events from detection to resolution.
3. **Root Cause:** The underlying technical cause (not "human error").
4. **Contributing Factors:** What made detection or resolution slower.
5. **Action Items:** Concrete, assigned tasks to prevent recurrence. Each item should reference a specific prevention measure (monitoring, test, architectural change).
6. **Lessons Learned:** What went well, what could improve in the response process.

Focus on systemic improvements, not blame. The goal is to make the system more resilient, not to find fault.
