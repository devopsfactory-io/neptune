# Governance

This document describes how the Neptune project is governed and how you can become a maintainer.

## Maintainers

Maintainers are listed in [MAINTAINERS.md](MAINTAINERS.md). They have write access to the repository and are responsible for:

- Reviewing and merging pull requests
- Triaging issues and guiding the roadmap
- Releasing versions and maintaining compatibility
- Upholding the [Code of Conduct](CODE_OF_CONDUCT.md) and [Security Policy](SECURITY.md)

## Adding maintainers

- New maintainers are proposed by existing maintainers (e.g. via a pull request that updates MAINTAINERS.md and optionally this file).
- The proposal should briefly state the nominee’s contributions and why they are a good fit.
- Existing maintainers reach consensus (e.g. lazy consensus over a short period; no formal vote required unless the group prefers it).
- Once agreed, a maintainer updates MAINTAINERS.md and the nominee is granted the necessary access.

We aim for maintainers from multiple organizations over time to keep the project neutral and sustainable.

## Removing maintainers

- A maintainer may step down at any time by opening a PR that moves them to the "Emeritus maintainers" section in MAINTAINERS.md.
- If a maintainer is inactive for a long period (e.g. no substantial participation for 6+ months), other maintainers may propose moving them to emeritus after a brief check-in (e.g. an issue or private message).
- Violations of the Code of Conduct or Security Policy are handled according to those documents and may result in removal; the CNCF Code of Conduct Committee may be involved for CoC matters.

## Decision making

- **Code and design**: Decisions are made through pull requests and issues. Maintainers review PRs and merge when there is consensus. Significant design or behavior changes should be discussed in an issue or PR description first.
- **Releases**: A maintainer cuts a release (e.g. by pushing a version tag) following the process in [AGENTS.md](AGENTS.md#ci). There is no formal release committee; maintainers coordinate as needed.
- **Disputes**: If maintainers disagree, they try to resolve it through discussion. If needed, they may involve a neutral third party (e.g. another maintainer or the CNCF, if the project is under the CNCF).

## Community

Everyone is welcome to contribute. See [CONTRIBUTING.md](CONTRIBUTING.md) for how to contribute and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for expected behavior.
