---
name: incident-debugging
description: "Systematic debugging workflow — reproduce, isolate, hypothesize, verify, fix, prevent. Includes log analysis, root cause table, bisect strategy, and post-incident review."
scope: universal
---

# Incident Debugging Skill

## When to Use

Invoke this skill when:
- Investigating a bug reported by users or QA
- Debugging a production incident (elevated error rate, degraded latency, outage)
- Analyzing error logs or alerting output
- Diagnosing intermittent or non-deterministic failures
- Conducting a post-incident review

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
  |
  +-- YES
  |     |
  |     +--> Capture exact reproduction steps (inputs, config, data state)
  |     |
  |     +--> Reproduce locally
  |     |
  |     +--> Isolate with binary search (Step 2)
  |     |
  |     +--> Fix (Step 5)
  |
  +-- NO
        |
        +--> Do you have logs or traces?
               |
               +-- YES
               |     |
               |     +--> Run log analysis (see below)
               |     |
               |     +--> Reconstruct timeline
               |     |
               |     +--> Form hypothesis from log evidence (Step 3)
               |     |
               |     +--> Attempt targeted reproduction based on hypothesis
               |
               +-- NO
                     |
                     +--> Add structured logging at suspect boundaries
                     |
                     +--> Add metrics for the suspected failure mode
                     |
                     +--> Set up alerts for recurrence
                     |
                     +--> Wait for recurrence with instrumentation in place
                     |
                     +--> Capture data on next occurrence --> restart from top
```

## Log Analysis Patterns

### Find the First Error

Cascading failures produce many log entries. The first error in the timeline is usually the root cause. Everything after is a consequence.

```
# Search for the first error in the time window
# Filter by timestamp range, then sort chronologically
# Look for ERROR or WARN level entries
```

### Correlate by Request ID / Trace ID

Use `request_id` or `trace_id` to follow a single request across services.

```
# Grep by request_id to see the full lifecycle of one request
# across all services and log files
```

### Timeline Reconstruction

1. Identify the time window: when did the issue start and stop?
2. Collect logs from all involved services for that window.
3. Sort by timestamp. Look for the transition from healthy to unhealthy.
4. Identify what changed at the transition point: deploy, config change, traffic spike, dependency failure.

### Pattern Matching

Look for:
- **Repeated error messages** with increasing frequency (cascading failure)
- **Gaps in expected log entries** (process crash, deadlock, or resource exhaustion)
- **Correlation with external events** (deploy timestamps, cron job schedules, traffic patterns)
- **Resource metrics at the time of failure** (CPU, memory, connection pool usage, queue depth)

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

### Manual Bisect

```
git bisect start
git bisect bad                    # current commit is broken
git bisect good <known-good-sha> # last known working commit

# Git checks out a middle commit. Test it.
git bisect good   # if this commit works
git bisect bad    # if this commit is broken

# Repeat until git identifies the first bad commit.
git bisect reset  # return to original branch when done
```

### Automated Bisect

Write a script that exits 0 for good and non-zero for bad:

```bash
#!/bin/bash
# bisect-test.sh
# Run the specific test that detects the bug
# Exit 0 = good, exit 1 = bad
go test ./internal/auth/ -run TestLoginFlow -count=1
```

```
git bisect start
git bisect bad HEAD
git bisect good <known-good-sha>
git bisect run ./bisect-test.sh
```

Git runs the script at each bisect step automatically and reports the first bad commit.

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
