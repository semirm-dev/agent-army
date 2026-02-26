<!-- Sync: Must stay in sync with cursor/504-testing.mdc -->

# 🧪 Testing Patterns

## Test Naming

- **Go:** `TestFunctionName_scenario` (e.g., `TestCreateUser_duplicateEmail`). Use `t.Run()` for subtests.
- **TypeScript:** `describe("FunctionName", () => { it("should do X when Y", ...) })`. Use clear, behavioral names.
- **Python:** `test_function_name_scenario` (e.g., `test_create_user_duplicate_email`). Use `class TestCreateUser:` for grouping.

## Table-Driven Test Structure

### Go

```go
func TestParseAmount(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int64
        wantErr bool
    }{
        {"valid amount", "42.50", 4250, false},
        {"negative amount", "-10.00", -1000, false},
        {"invalid format", "abc", 0, true},
        {"empty string", "", 0, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseAmount(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseAmount(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ParseAmount(%q) = %v, want %v", tt.input, got, tt.want)
            }
        })
    }
}
```

### TypeScript

```typescript
describe("parseAmount", () => {
  const cases = [
    { name: "valid amount", input: "42.50", expected: 4250 },
    { name: "negative amount", input: "-10.00", expected: -1000 },
    { name: "invalid format", input: "abc", expectedError: ParseError },
    { name: "empty string", input: "", expectedError: ParseError },
  ] as const;

  for (const { name, input, expected, expectedError } of cases) {
    it(name, () => {
      if (expectedError) {
        expect(() => parseAmount(input)).toThrow(expectedError);
      } else {
        expect(parseAmount(input)).toBe(expected);
      }
    });
  }
});
```

### Python

```python
@pytest.mark.parametrize(
    "input_val, expected, raises",
    [
        ("42.50", 4250, None),
        ("-10.00", -1000, None),
        ("abc", None, ParseError),
        ("", None, ParseError),
    ],
)
def test_parse_amount(input_val: str, expected: int | None, raises: type[Exception] | None) -> None:
    if raises:
        with pytest.raises(raises):
            parse_amount(input_val)
    else:
        assert parse_amount(input_val) == expected
```

## Fixture and Factory Patterns

### Builder Pattern for Test Data

Prefer factories over raw data construction:
- Create sensible defaults for all required fields
- Allow overrides for the fields relevant to each test
- Keep test data close to the test (in the same file or a `testutil` package)

### Test Isolation

- **Database tests:** Wrap each test in a transaction, rollback after. Use `t.Cleanup()` (Go), `afterEach` (TS), `yield` fixtures (Python).
- **File system tests:** Use temp directories. Clean up with `t.TempDir()` (Go), `tmp` (TS), `tmp_path` fixture (Python).
- **In-memory databases:** Use SQLite in-memory for fast unit tests when full DB features aren't needed.
- **Network isolation:** No real HTTP calls in unit tests. Use fakes, MSW, or recorded responses.

## Flaky Test Prevention

- **No `time.Sleep` / `setTimeout` for synchronization.** Use polling, waitFor, or channels.
- **No network calls in unit tests.** Mock at the boundary.
- **Deterministic test data.** Use factories with fixed values, not `Math.random()` or `time.Now()`.
- **Isolated test state.** Each test creates its own data. No shared mutable fixtures.
- **Explicit timeouts.** Set test timeouts to catch hangs: `go test -timeout 30s`, `jest --testTimeout 10000`.

## CI Parallelization

- **Go:** `go test -parallel N` controls per-test parallelism. `-count=1` disables test caching. Use `-race` always.
- **TypeScript:** vitest: `--pool threads` (default), `--pool forks` for isolation. jest: `--maxWorkers=N`.
- **Python:** `pytest-xdist`: `-n auto` for auto-detected parallelism. Use `--dist loadfile` to keep test files together.

## Coverage Reporting

- Run coverage as part of CI, not just locally
- Set coverage thresholds as CI gates (fail build if coverage drops)
- **Critical paths** (auth, payments, data mutations): 80%+ line coverage
- **Utilities and shared libraries:** 90%+ line coverage
- **Generated code:** No coverage requirement
- Prefer branch coverage over line coverage where tools support it
