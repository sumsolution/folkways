```
fix(auth): treat tokens near expiry as expired

isExpired compared expiresAt to the current time with no margin, so a
request that started just before expiry could fail partway through
with a 401. Treat a token as expired once it is within CLOCK_SKEW_MS
(30s) of expiresAt, so it is refreshed before the request starts.

User-visible effect: fewer intermittent 401s and forced re-logins for
requests made right at the end of a token's lifetime.

Add tests for a token inside and outside the skew window.

Closes EX-142

```

I used the format from your personal overlay (`fix(auth):` header, `Closes EX-142` footer). The team overlay still says `auth: Summary` with `Refs: EX-142`, so switch to that if the team format hasn't changed yet.
