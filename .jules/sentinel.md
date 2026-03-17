## 2024-05-18 - [Overly Permissive CORS Configuration]
**Vulnerability:** The API allowed `Access-Control-Allow-Origin: *` while simultaneously setting `Access-Control-Allow-Credentials: true`. This combination is a security risk as it allows any origin to make authenticated cross-origin requests using the user's credentials, potentially exposing sensitive data.
**Learning:** Hardcoding `*` for allowed origins alongside credentials allows for CSRF-like attacks where a malicious site can read responses using a user's session.
**Prevention:** Implement a dynamic origin validation mechanism that checks incoming origins against an explicit, configurable allowlist before setting the `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials` headers.

## 2026-03-13 - [Wildcard CORS + reflect Origin + credentials]
**Vulnerability:** Treating `ALLOWED_ORIGINS=*` by reflecting the request `Origin` and setting `Access-Control-Allow-Credentials: true` is worse than `*` + credentials (browsers reject the latter). Reflected origin + credentials is accepted, so any site could perform credentialed cross-origin requests.
**Prevention:** If open CORS is required, use literal `Access-Control-Allow-Origin: *` and omit `Access-Control-Allow-Credentials`. For cookies/auth cross-origin, use an explicit origin allowlist and reflect only listed origins.

## 2026-03-17 - [GORM Negative Limit DoS Vulnerability]
**Vulnerability:** GORM translates negative limit values (like `Limit(-1)`) to 'no limit'. If a user supplies a negative limit parameter (e.g., `?limit=-1`), the application fetches all records from the database instead of paginating, causing a Denial of Service (DoS) and memory exhaustion.
**Learning:** Default parameter parsing does not protect against negative values, and GORM's behavior with negative limits creates an unexpected security gap in data fetching endpoints.
**Prevention:** Always strictly validate and bound pagination parameters before passing them to GORM. Ensure limits are strictly positive (`> 0`) and capped to a sensible maximum (e.g., 1000) to prevent excessive data retrieval.
