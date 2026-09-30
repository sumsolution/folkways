```
auth: Treat tokens near expiry as expired

A request that started just before a token expired could be rejected
with a 401 partway through. isExpired now treats a token as expired
once it is within CLOCK_SKEW_MS (30s) of expiresAt, so callers refresh
before the request starts instead of failing during it.

The comparison is now inclusive (<=), so a token is expired exactly at
the skew boundary. Tests cover a token inside and outside the skew
window.

User-visible effect: fewer spurious 401s and forced re-logins for
requests made close to token expiry.

Refs: EX-142
```

I followed the team overlay and left the ticket key out of the header. Your personal overlay asks for `[EX-142]` in the header and no `Refs:` trailer, but the team overlay forbids a ticket key in the header and requires the trailer. If you'd rather use your format, swap the header to `[EX-142] Treat tokens near expiry as expired` and drop the last line.
