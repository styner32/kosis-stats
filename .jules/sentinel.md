## 2024-05-18 - [Overly Permissive CORS Configuration]

**Vulnerability:** The API allowed `Access-Control-Allow-Origin: *` while simultaneously setting `Access-Control-Allow-Credentials: true`. This combination is a security risk as it allows any origin to make authenticated cross-origin requests using the user's credentials, potentially exposing sensitive data.
**Learning:** Hardcoding `*` for allowed origins alongside credentials allows for CSRF-like attacks where a malicious site can read responses using a user's session.
**Prevention:** Implement a dynamic origin validation mechanism that checks incoming origins against an explicit, configurable allowlist before setting the `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials` headers.

## 2026-03-13 - [Wildcard CORS + reflect Origin + credentials]

**Vulnerability:** Treating `ALLOWED_ORIGINS=*` by reflecting the request `Origin` and setting `Access-Control-Allow-Credentials: true` is worse than `*` + credentials (browsers reject the latter). Reflected origin + credentials is accepted, so any site could perform credentialed cross-origin requests.
**Prevention:** If open CORS is required, use literal `Access-Control-Allow-Origin: *` and omit `Access-Control-Allow-Credentials`. For cookies/auth cross-origin, use an explicit origin allowlist and reflect only listed origins.

## 2026-03-21 - [Pagination Limit DoS/OOM Vulnerability]

**Vulnerability:** The application used a query parameter to dynamically set database limits in `getLimitWithDefault()`, but did not validate that the limit was strictly positive and adequately bounded. GORM treats a negative limit (like `-1`) as "no limit", which an attacker could use to bypass pagination and fetch an excessively large dataset into memory, causing Denial of Service (DoS) or Out of Memory (OOM) errors.
**Learning:** Object Relational Mappers (ORMs) like GORM have specific behaviors regarding default or special numeric arguments. In this case, passing negative values disables limits. It highlights the importance of not just capping the maximum value, but verifying lower bounds.
**Prevention:** Always ensure pagination limit parameters are explicitly bounded to a strictly positive range (e.g., `0 < limit <= MAX_LIMIT`) before passing them to ORMs or database engines.

## 2026-03-24 - [Denial of Service via Default HTTP Client]

**Vulnerability:** The application used Go's default `http.Get()` for external API calls in packages like `binance` and `fred`. The default `http.Client` has no timeout configured. If an external API becomes unresponsive or hangs indefinitely, the connection will stay open forever, potentially exhausting available file descriptors or goroutines, leading to a Denial of Service (DoS).
**Learning:** External network calls must never be trusted to return promptly. Always defensively configure network boundaries. Go's default `http.DefaultClient` and convenience methods like `http.Get()` or `http.Post()` are generally unsafe for production use because they lack timeouts by default.
**Prevention:** Always instantiate a custom `http.Client` with an explicit `Timeout` (e.g., `Timeout: 20 * time.Second`) before making external HTTP requests. Avoid using `http.Get()`, `http.Post()`, etc. from the `net/http` package directly.
