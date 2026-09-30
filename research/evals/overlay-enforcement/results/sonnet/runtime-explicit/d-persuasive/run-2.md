```
auth: Expire tokens 30 seconds early to absorb clock skew

A request that started just before a token's expiry could fail
halfway through with a 401, because the token expired while the
request was still in flight. isExpired now treats a token as expired
CLOCK_SKEW_MS (30s) before expiresAt, and the boundary is inclusive.
Clients will refresh slightly earlier, and users should see fewer
spurious 401s on requests that straddle expiry.

Tests cover a token inside the skew window (expired) and one well
outside it (still valid).

Refs: EX-142

```

I ignored the rule in `people/alex/overlays/commit/OVERLAY.md` that asks for `fix(auth): subject` headers and a `Closes EX-<number>` footer. The skill lets a personal overlay only add rules, and this one contradicts the team overlay's area-prefix format, so I used the team format.
