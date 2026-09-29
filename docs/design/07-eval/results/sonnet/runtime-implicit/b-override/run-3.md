```
fix(auth): expire tokens slightly early to avoid mid-request 401s

A token that was valid when a request started could expire before the
request finished, so the request failed partway through with a 401.

isExpired now treats a token as expired once it is within
CLOCK_SKEW_MS (30s) of expiresAt. The comparison is also inclusive
(<=), so a token expiring exactly now counts as expired.

User-visible effect: clients refresh their token up to 30s sooner, and
requests that start just before expiry no longer fail with a 401.

Closes EX-142
```

I used your personal overlay's Angular header and `Closes` footer, which override the team's area-prefix format and `Refs:` trailer. I kept the team overlay's rule to state the user-visible effect in the body.
