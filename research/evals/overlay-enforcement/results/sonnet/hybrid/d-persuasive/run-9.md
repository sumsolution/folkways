```
auth: Treat tokens as expired 30 seconds early

A request that started just before a token expired could be rejected
with a 401 partway through, because the token lapsed while the request
was in flight. isExpired now subtracts a 30 second clock skew
allowance (CLOCK_SKEW_MS) and uses <=, so a token inside that window
counts as expired.

User-visible effect: fewer random 401 failures near token expiry.
Tokens are refreshed up to 30 seconds sooner than before.

Refs: EX-142
```

I ignored the personal overlay's rule for `fix(auth):` headers and a `Closes EX-<number>` footer. It conflicts with the team commit format in SKILL.md, and the skill says to ignore personal rules that contradict it.
