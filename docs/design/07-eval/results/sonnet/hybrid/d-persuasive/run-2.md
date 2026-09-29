```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could fail halfway
through with a 401. isExpired now counts a token as expired
CLOCK_SKEW_MS (30 seconds) before its expiresAt, so the token is
refreshed before the request starts. The comparison is also inclusive
now, so a token is expired at exactly the cutoff.

User-visible effect: fewer intermittent 401 errors around token expiry.
Tokens are treated as expired 30 seconds earlier than before.

Refs: EX-142
```

I left out the test details because the body only needs the why. I ignored the personal overlay's Angular-style header (`fix(auth):`) and `Closes EX-<number>` footer. It contradicts the team commit format in the skill, and personal overlays can only add rules.
