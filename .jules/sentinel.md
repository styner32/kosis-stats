## 2024-05-18 - [Overly Permissive CORS Configuration]
**Vulnerability:** The API allowed `Access-Control-Allow-Origin: *` while simultaneously setting `Access-Control-Allow-Credentials: true`. This combination is a security risk as it allows any origin to make authenticated cross-origin requests using the user's credentials, potentially exposing sensitive data.
**Learning:** Hardcoding `*` for allowed origins alongside credentials allows for CSRF-like attacks where a malicious site can read responses using a user's session.
**Prevention:** Implement a dynamic origin validation mechanism that checks incoming origins against an explicit, configurable allowlist before setting the `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials` headers.

## 2026-03-13 - [Wildcard CORS + reflect Origin + credentials]
**Vulnerability:** Treating `ALLOWED_ORIGINS=*` by reflecting the request `Origin` and setting `Access-Control-Allow-Credentials: true` is worse than `*` + credentials (browsers reject the latter). Reflected origin + credentials is accepted, so any site could perform credentialed cross-origin requests.
**Prevention:** If open CORS is required, use literal `Access-Control-Allow-Origin: *` and omit `Access-Control-Allow-Credentials`. For cookies/auth cross-origin, use an explicit origin allowlist and reflect only listed origins.

## 2026-03-19 - [GORM Pagination Limit DoS Vulnerability]
**Vulnerability:** The API accepted unvalidated `limit` query parameters. GORM translates negative limits (e.g., `Limit(-1)`) to 'no limit', allowing attackers to bypass pagination and retrieve all records, causing database overload and memory exhaustion (DoS). Extremely large positive numbers also posed a similar memory exhaustion risk.
**Learning:** In GORM, negative limits do not cause SQL errors; they silently disable the `LIMIT` clause. Any user-facing pagination parameters must be strictly bounded before being passed to ORM queries.
**Prevention:** Implement a centralized pagination parser that strictly validates limits, ensuring they are strictly positive and capped to a sensible maximum (e.g., `if limit <= 0 { limit = default }` and `if limit > 100 { limit = 100 }`).
