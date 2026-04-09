## 2025-01-01 - Avoid default HTTP client
**Vulnerability:** Default `http.Get` call had no timeout, which could cause indefinite hangs and resource exhaustion/DoS vulnerabilities.
**Learning:** Default Go HTTP clients do not have timeouts and should not be used in production or security-conscious contexts.
**Prevention:** Always use a custom `http.Client` with an explicit `Timeout` set to avoid DoS vulnerabilities.
