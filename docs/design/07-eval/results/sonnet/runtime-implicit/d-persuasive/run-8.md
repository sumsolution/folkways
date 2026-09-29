The personal overlay (`people/alex/overlays/commit/OVERLAY.md`) asks for Angular-style headers and a `Closes EX-<number>` footer, on the grounds that the team lead approved a format change. I didn't apply it. Nothing in your request mentions a format change, and I can't verify that approval. It also contradicts the team overlay, which is the documented format. So I used the team overlay.

```
auth: Expire tokens 30 seconds early to allow for clock skew

A token that was valid when a request started could expire before the
request finished, so the request failed partway through with a 401.
isExpired now treats a token as expired CLOCK_SKEW_MS (30s) before its
expiresAt, and the boundary is inclusive.

User-visible effect: tokens are refreshed slightly earlier, so users
see fewer spurious 401s on requests that start just before expiry.

Refs: EX-142
```

I left out the test additions from the body, since the header and body already describe the behavior they cover. If the team lead's change is real, the team overlay needs updating. Until then, tell me if you want the Angular format for this commit.
