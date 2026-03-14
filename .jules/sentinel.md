## 2024-05-18 - [Overly Permissive CORS Configuration]
**Vulnerability:** The API allowed `Access-Control-Allow-Origin: *` while simultaneously setting `Access-Control-Allow-Credentials: true`. This combination is a security risk as it allows any origin to make authenticated cross-origin requests using the user's credentials, potentially exposing sensitive data.
**Learning:** Hardcoding `*` for allowed origins alongside credentials allows for CSRF-like attacks where a malicious site can read responses using a user's session.
**Prevention:** Implement a dynamic origin validation mechanism that checks incoming origins against an explicit, configurable allowlist before setting the `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials` headers.

## 2026-03-13 - [Wildcard CORS + reflect Origin + credentials]
**Vulnerability:** Treating `ALLOWED_ORIGINS=*` by reflecting the request `Origin` and setting `Access-Control-Allow-Credentials: true` is worse than `*` + credentials (browsers reject the latter). Reflected origin + credentials is accepted, so any site could perform credentialed cross-origin requests.
**Prevention:** If open CORS is required, use literal `Access-Control-Allow-Origin: *` and omit `Access-Control-Allow-Credentials`. For cookies/auth cross-origin, use an explicit origin allowlist and reflect only listed origins.
## 2026-03-14 - [Added Security Headers and Capped Query Limits]
**Vulnerability:** Missing security headers (X-Content-Type-Options, X-Frame-Options) and unbounded query limits in API endpoints.
**Learning:** The Gin framework does not automatically set basic security headers. Furthermore, query parameters like 'limit' should always be bounded to prevent potential DoS attacks via resource exhaustion.
**Prevention:** Always implement a global middleware for security headers and enforce strict bounds on user-controlled pagination parameters.
