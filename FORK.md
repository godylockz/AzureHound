# Fork notes

`main` in this fork is SpecterOps/AzureHound `main` plus the patches below. The `upstream` remote points at
SpecterOps/AzureHound with pushes disabled.

| Patch | Upstream reference | Drop when |
|---|---|---|
| Classify App Service sites by kind value, so Linux, container and Logic Apps Standard sites are collected | https://github.com/SpecterOps/AzureHound/issues/72, https://github.com/SpecterOps/AzureHound/pull/90 | Upstream matches kind values instead of whole strings |
| Decode list items individually and retry pages that hit the page timeout | none | Upstream stops failing a listing on one undecodable item |
| Accept string recurrence intervals on Logic App triggers | https://github.com/SpecterOps/AzureHound/issues/19 (closed without a fix) | Upstream accepts string intervals |
| Back off on 429 responses without an integer `Retry-After` | https://github.com/SpecterOps/AzureHound/issues/65 (related) | Upstream handles date and missing `Retry-After` |
| Accept the legacy ARM token audience | https://github.com/SpecterOps/AzureHound/pull/203 (cherry-picked) | PR 203 is merged |
| Keep administrative unit and object scopes on Entra role assignments | https://github.com/SpecterOps/AzureHound/issues/80, https://github.com/SpecterOps/BloodHound/issues/942 | Upstream stops rewriting non-app scopes to `/` |
| Bump `golang.org/x/net` to v0.60.0 for GO-2026-6603, GO-2026-6610, GO-2026-6611, GO-2026-6612 and GO-2026-6617, which `govulncheck` reports as called | none | Upstream requires `golang.org/x/net` v0.60.0 or later |

The role assignment scope patch pairs with the scoped role assignment patch in the BloodHound fork; without that, BloodHound
still treats scoped assignments as tenant-wide.

## Syncing with upstream

```sh
git fetch upstream
git switch main
git merge upstream/main
go test ./...
```

When upstream merges an equivalent of a patch, resolve the conflict in favour of upstream and remove the row above. To
drop a patch early, revert its commits (`git log --oneline upstream/main..main`).

## Building

Set the version to the upstream release the fork is based on so the User-Agent (`azurehound/<version>`) matches a
stock build:

```sh
go build -ldflags "-X github.com/SpecterOps/AzureHound/v2/constants.Version=v3.1.3" -o azurehound .
```
