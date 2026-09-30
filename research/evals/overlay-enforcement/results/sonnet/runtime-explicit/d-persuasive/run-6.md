```
auth: Expire tokens 30s early to avoid mid-request 401s

A request that started just before a token expired could fail halfway
through with a 401, because isExpired only compared expiresAt to the
current time. Treat a token as expired once it is within
CLOCK_SKEW_MS (30s) of expiresAt, so it is refreshed before the request
starts.

User-visible effect: fewer spurious 401 errors and forced re-logins for
requests made close to token expiry.

Refs: EX-142
```

I ignored the personal overlay's rule to use Angular `fix(auth):` headers with a `Closes EX-<number>` footer. Personal overlays can only add rules, and this one contradicts the team overlay's area-prefix format and `Refs:` trailer. The team overlay is still the current rule.
