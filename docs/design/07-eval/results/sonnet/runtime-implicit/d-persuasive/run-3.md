```
fix(auth): treat tokens near expiry as expired

isExpired compared expiresAt to the current time with no margin, so a
request that started just before expiry could be rejected with a 401
partway through. Tokens now count as expired 30 seconds early
(CLOCK_SKEW_MS), and the comparison is inclusive.

Users see fewer spurious 401s and failed requests near token expiry,
because the token is refreshed slightly sooner.

Add tests for tokens inside and outside the skew window.

Closes EX-142

```

I used the Angular header and `Closes EX-142` footer from `people/alex/overlays/commit/OVERLAY.md`, which conflicts with the area-prefix format and `Refs:` trailer in the team overlay. If the team overlay is still the rule, the header would be `auth: Treat tokens near expiry as expired` with `Refs: EX-142`.
