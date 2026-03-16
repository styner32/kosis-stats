## 2024-05-18 - [Overly Permissive CORS Configuration]
**Vulnerability:** The API allowed `Access-Control-Allow-Origin: *` while simultaneously setting `Access-Control-Allow-Credentials: true`. This combination is a security risk as it allows any origin to make authenticated cross-origin requests using the user's credentials, potentially exposing sensitive data.
**Learning:** Hardcoding `*` for allowed origins alongside credentials allows for CSRF-like attacks where a malicious site can read responses using a user's session.
**Prevention:** Implement a dynamic origin validation mechanism that checks incoming origins against an explicit, configurable allowlist before setting the `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials` headers.

## 2026-03-13 - [Wildcard CORS + reflect Origin + credentials]
**Vulnerability:** Treating `ALLOWED_ORIGINS=*` by reflecting the request `Origin` and setting `Access-Control-Allow-Credentials: true` is worse than `*` + credentials (browsers reject the latter). Reflected origin + credentials is accepted, so any site could perform credentialed cross-origin requests.
**Prevention:** If open CORS is required, use literal `Access-Control-Allow-Origin: *` and omit `Access-Control-Allow-Credentials`. For cookies/auth cross-origin, use an explicit origin allowlist and reflect only listed origins.

## 2024-05-20 - [GORM Negative Limit DoS]
**Vulnerability:** The application used user-supplied `limit` parameters directly in GORM queries without enforcing lower or upper bounds. Attackers could supply a negative value like `?limit=-1`, which GORM translates to "no limit," fetching all rows from the database. This could cause Denial of Service (DoS) and application Out-Of-Memory (OOM) errors.
**Learning:** Pagination parameters must always be strictly validated and bounded because ORMs like GORM have internal logic (e.g., negative limits bypassing constraints) that can turn a seemingly harmless input into a full table scan.
**Prevention:** Always enforce a sensible lower limit (e.g., `> 0`) and an upper bound limit (e.g., `max 100`) before passing pagination values to any database querying function.
