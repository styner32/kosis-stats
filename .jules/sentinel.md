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

## 2026-04-06 - [HTTP Client Timeout Bypass Vulnerability]

**Vulnerability:** The application was making external API calls using the default `http.Get()`, which does not enforce any timeout. Even when a custom `http.Client` with an explicit timeout was configured within a struct (like `FREDClient`), the code incorrectly bypassed it and still used `http.Get()`. This lack of a timeout can lead to hanging connections and potential Denial of Service (DoS) or resource exhaustion vulnerabilities if the external server is unresponsive or slow.
**Learning:** Developers sometimes configure an `http.Client` but forget to use it for actual requests, falling back to standard `http` package functions out of habit. Default Go HTTP clients lack timeouts, making them unsafe for production code interacting with external APIs.
**Prevention:** Always use the explicitly configured `http.Client` (e.g., `c.client.Get(url)`) for all outgoing HTTP requests. Never use the default `http.Get()`, `http.Post()`, etc., without explicitly setting timeouts on the `http.DefaultClient`, but prefer independent client instances to prevent global state issues. Review external API integrations specifically for timeout configurations.
