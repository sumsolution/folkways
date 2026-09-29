```
fix(auth): treat tokens near expiry as expired

isExpired compared expiresAt to now exactly, so a request that began
just before expiry could be sent with a token that expired mid-flight
and fail with a 401. Treat a token as expired once it is within
CLOCK_SKEW_MS (30s) of expiresAt, and export the constant.

User-visible effect: users no longer hit sporadic 401s on requests
started right before their token expires. The token is refreshed a
little earlier instead.

Add tests for tokens inside and outside the skew window.

Closes EX-142

```

I used your personal overlay's Angular header and `Closes EX-142` footer instead of the team overlay's area-prefix header and `Refs:` trailer. I kept the team overlay's request to describe the user-visible effect in the body.
