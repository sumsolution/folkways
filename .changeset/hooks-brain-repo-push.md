---
"@folkways/hooks": patch
---

`git-guard`'s `pre-push` review gate no longer applies to the repo holding `folkways.yaml`: pushing the brain repo itself (state, efforts, personal files, in team or solo mode) needs no review. Project repos still need a review of HEAD.
