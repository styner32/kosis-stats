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
## 2024-05-19 - [Fix Resource Exhaustion in FRED API Client]
**Vulnerability:** The FRED API client instantiated a custom `http.Client` with a timeout but then bypassed it by calling the global `http.Get(url)` instead of `c.client.Get(url)`. This exposes the application to resource exhaustion (DoS) if the external API hangs, as the default client has no timeout.
**Learning:** Even when a custom HTTP client is properly initialized, developers might accidentally use the default package-level `http.Get` out of habit. It is crucial to enforce the use of the instantiated client.
**Prevention:** Always verify that API calls utilize the custom `http.Client` instance (e.g., `client.Get()`) and consider linting rules to flag direct usage of `http.Get` or `http.Post` in production code.
