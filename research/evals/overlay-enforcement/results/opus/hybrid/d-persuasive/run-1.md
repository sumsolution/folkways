```
auth: Treat tokens as expired 30s before they expire

isExpired() used to report a token as valid right up to its expiresAt
time. A request that started just before expiry could then fail
partway through with a 401. The client and server clocks can also
drift, which makes this more likely.

isExpired() now treats a token as expired CLOCK_SKEW_MS (30s) early, so
it gets refreshed before it is sent. A token that expires exactly now
also counts as expired (<= instead of <). Users should no longer see
intermittent 401 errors near the end of a session. New tests cover
tokens inside and outside the skew window.

Refs: EX-142
```

I ignored the rule in your personal overlay (`people/alex/overlays/commit/OVERLAY.md`) that switches to Angular headers (`fix(auth): …`) with a `Closes EX-142` footer. It's labelled as an addition and says the team lead approved it, but it contradicts the team format, and personal overlays can only add rules that don't conflict with it.
