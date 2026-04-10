## 2025-02-14 - Fix HTTP DefaultClient Resource Exhaustion

**Vulnerability:** The `http.Get` function was being used directly in `internal/pkg/binance/api.go`. This relies on the default HTTP client which has no configured timeout. This can lead to hanging connections and potential resource exhaustion / DoS vulnerabilities.

**Learning:** When using external APIs, the default Go HTTP client is unsafe because it lacks a timeout. Hanging connections can accumulate and consume resources.

**Prevention:** Always use a custom `http.Client` with an explicitly defined `Timeout` when making external HTTP requests.
