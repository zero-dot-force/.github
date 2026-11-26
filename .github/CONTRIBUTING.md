# Contributing to Zero Dot Force

Thanks for helping out. Zero Dot Force is the junior hero team of
[Unbound Force](https://unboundforce.dev): new tools start here and team up
with the senior heroes before they're ready for the big leagues. This guide
applies to every repo in the org unless a repo has its own `CONTRIBUTING.md`.

By taking part, you agree to follow our
[Code of Conduct](CODE_OF_CONDUCT.md).

## Where to start

| Repo | What it is |
| ---- | ---------- |
| [vibe-check](https://github.com/zero-dot-force/vibe-check) | Coupling, cohesion and dependency-cycle metrics |
| [reading-stone](https://github.com/zero-dot-force/reading-stone) | TypeScript analyzer for Gaze |
| [snake-eyes](https://github.com/zero-dot-force/snake-eyes) | Python analyzer for Gaze |
| [.github](https://github.com/zero-dot-force/.github) | Org settings, teams and these files |

Each repo's `README.md` covers setup, and its `AGENTS.md`, where present,
describes the workflow for AI coding agents. Issues labelled
[good first issue](https://github.com/search?q=org%3Azero-dot-force+label%3A%22good+first+issue%22+state%3Aopen&type=issues)
are a good way in.

## Reporting bugs and asking for features

Open an issue on the repo it affects. For a bug, include:

- the version (`--version`, or the package version you installed)
- the exact command or call you ran
- what you expected and what happened instead

For a feature, describe the problem you're trying to solve before the
solution you have in mind.

Security problems don't belong in public issues. See
[SECURITY.md](SECURITY.md).

## Making a change

1. For anything beyond a small fix, open an issue first so we can agree on
   the approach before you write the code.
2. Fork the repo and create a branch from `main`.
3. Make the change, with tests. Repos with an `openspec/` directory track
   non-trivial changes as an OpenSpec change. Follow that repo's
   `AGENTS.md` for how.
4. Write commit messages in the
   [Conventional Commits](https://www.conventionalcommits.org/) format, for
   example `fix: handle empty files in the TypeScript adapter`. Common types
   are `feat`, `fix`, `docs`, `test`, `refactor`, `ci` and `chore`.
5. Open a pull request against `main` and fill in the template.

## Before a pull request can merge

The default branch of each code repo is protected:

- The repo's required CI checks must pass.
- At least one maintainer must approve, and a new push after that
  approval needs approving again.
- Force pushes to `main` are blocked.

Maintainers may ask for changes. Push them to the same branch, and the pull
request updates.

## Licensing

Every repo is licensed under the Apache License 2.0. By contributing, you
agree that your contributions are licensed under the same terms.
