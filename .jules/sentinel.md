## 2024-05-18 - [Overly Permissive CORS Configuration]
**Vulnerability:** The API allowed `Access-Control-Allow-Origin: "*"` combined with `Access-Control-Allow-Credentials: "true"` in the global Gin middleware.
**Learning:** This is a severe security risk (CORS misconfiguration) that could allow malicious websites to perform Cross-Site Request Forgery (CSRF) via credentialed requests, compromising user data or actions. Modern browsers typically reject this combination, but relying on browser behavior is insufficient defense.
**Prevention:** Implement strict Origin validation by checking the incoming `Origin` header against a whitelist of allowed domains loaded from configuration. Do not blindly reflect origins or use wildcard `*` when credentials are permitted.
