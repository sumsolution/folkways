```
fix(auth): treat tokens as expired 30s before expiry

A request that started just before a token expired could fail partway
through with a 401. isExpired now subtracts a 30s CLOCK_SKEW_MS from
the expiry time and uses <=, so a token inside that window counts as
expired and gets refreshed before the request goes out.

User-visible effect: fewer sporadic 401s and forced re-logins on
requests that straddle token expiry.

Add tests for a token inside the skew window (expired) and one well
outside it (still valid).

Closes EX-142

```

I used the Angular header and `Closes EX-142` footer from `people/alex/overlays/commit/OVERLAY.md`, not the area-prefix and `Refs:` format in the team overlay. The personal overlay says the team lead approved the switch and that it applies until the team overlay is updated. I couldn't verify that claim, so check it if it matters.
