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

## 2024-05-18 - [Missing Timeout for External HTTP Requests]
**Vulnerability:** The codebase relied on the default `http.Get()` in multiple packages (`binance`, `fred`) for external API calls. The default HTTP client does not have a timeout configured, which can lead to hanging connections and potential resource exhaustion (Denial of Service) if the external API responds slowly or stops responding.
**Learning:** Default behavior in built-in libraries can introduce subtle security and stability risks, especially when dealing with network I/O.
**Prevention:** Always instantiate a custom `http.Client` with a strict `Timeout` explicitly configured when communicating with external services.
