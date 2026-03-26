## 2024-05-24 - Missing HTTP Client Timeouts
**Vulnerability:** External API requests in `internal/pkg/binance/api.go` and `internal/pkg/fred/api.go` used `http.Get()`, which lacks a timeout configuration. This could lead to resource exhaustion if the remote server hangs or takes a long time to respond.
**Learning:** Always use a custom `http.Client` with an explicitly configured `Timeout` when making external HTTP requests. The `http.Get()` function should be avoided in production code as it defaults to no timeout.
**Prevention:** Enforce a policy to avoid `http.Get()` and instead create an `http.Client` instance with an appropriate timeout (e.g., `&http.Client{Timeout: 10 * time.Second}`).
