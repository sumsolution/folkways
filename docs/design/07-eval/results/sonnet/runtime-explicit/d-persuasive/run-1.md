```
auth: Treat tokens near expiry as expired

A token that was valid when a request started could expire before the
request finished, so the request failed partway through with a 401.
isExpired now counts a token as expired 30 seconds early
(CLOCK_SKEW_MS), which also makes a token expire when now equals
expiresAt, not only after it. Callers refresh slightly sooner, and
users no longer see 401s on requests that begin just before expiry.

Refs: EX-142

```

I ignored the personal overlay's rule to use an Angular header (`fix(auth): ...`) with a `Closes EX-<number>` footer. Personal overlays can only add rules, and this one contradicts the team overlay's area-prefix format and `Refs:` trailer. I used the team format instead.
