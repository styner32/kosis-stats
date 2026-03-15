## 2024-05-18 - [Overly Permissive CORS Configuration]
**Vulnerability:** The API allowed `Access-Control-Allow-Origin: *` while simultaneously setting `Access-Control-Allow-Credentials: true`. This combination is a security risk as it allows any origin to make authenticated cross-origin requests using the user's credentials, potentially exposing sensitive data.
**Learning:** Hardcoding `*` for allowed origins alongside credentials allows for CSRF-like attacks where a malicious site can read responses using a user's session.
**Prevention:** Implement a dynamic origin validation mechanism that checks incoming origins against an explicit, configurable allowlist before setting the `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials` headers.

## 2026-03-13 - [Wildcard CORS + reflect Origin + credentials]
**Vulnerability:** Treating `ALLOWED_ORIGINS=*` by reflecting the request `Origin` and setting `Access-Control-Allow-Credentials: true` is worse than `*` + credentials (browsers reject the latter). Reflected origin + credentials is accepted, so any site could perform credentialed cross-origin requests.
**Prevention:** If open CORS is required, use literal `Access-Control-Allow-Origin: *` and omit `Access-Control-Allow-Credentials`. For cookies/auth cross-origin, use an explicit origin allowlist and reflect only listed origins.

## 2026-03-14 - [GORM Negative Limit DoS Vulnerability]
**Vulnerability:** GORM translates negative limit values (like `Limit(-1)`) to "no limit". When an API exposes pagination limits derived from user input without validation, attackers can request negative limits to force the database to fetch all records, causing memory exhaustion and Denial of Service (DoS).
**Learning:** Even strong typing doesn't prevent logic abuse. GORM's API design makes it easy to accidentally allow unbounded queries if user inputs for limits aren't explicitly bounded on both ends.
**Prevention:** Always validate integer inputs used for database query limits to ensure they are strictly positive and bounded by a reasonable maximum limit before passing them to the database layer (e.g., `if limit <= 0 || limit > MAX_ALLOWED`).
