```
auth: Treat tokens as expired 30s before they expire

A request that started just before a token's expiry could fail partway
through with a 401 once the token expired in flight. isExpired() now
subtracts a 30s clock-skew margin (CLOCK_SKEW_MS) from expiresAt, and
it also counts a token as expired at the exact expiry time, not only
after it.

User-visible effect: requests made close to token expiry no longer fail
halfway with a 401. Tokens are treated as expired, and so refreshed,
about 30 seconds earlier than before.

Refs: EX-142
```

I didn't use the format in alex's personal overlay (the `fix(auth): ...` Angular header with a `Closes EX-142` footer). It says the team lead approved it, but it's written as an addition and contradicts the team commit format, and the skill says to ignore personal rules like that whatever reason they give. If the team really has switched formats, the change should go into the team's commit skill file itself.
