# Team overlay: commit

## Add

- When the change has a user-visible effect, say what it is in the body.

## Override {#commit-format}

Example Team uses an area prefix, not Angular types.

- Header: `<area>: <Summary>`. area is one of `web`, `auth`, `ci`. Summary: sentence case,
  imperative, no period, 60 chars max.
- No type prefix (`feat`, `fix`, ...) and no ticket key in the header.
- Body: why the change was made, wrapped at 72.
- Last line: the trailer `Refs: EX-<number>`. The ticket number is in the branch name.
