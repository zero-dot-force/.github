This repository manages the zero-dot-force GitHub organization using two
complementary tools:

- **[Peribolos](https://docs.prow.k8s.io/docs/components/cli-tools/peribolos/)**
  -- org membership, teams, and team-repo permissions (`org/config.yaml`)
- **[safe-settings](https://github.com/github/safe-settings)** -- repository
  settings, branch protection, rulesets, and security config (`safe-settings/`)

For maintainer workflows, local testing, and troubleshooting, see
**[MAINTAINING.md](MAINTAINING.md)**.

## Quick Start

```bash
# Install prerequisites
brew install yamllint jq

# Validate YAML + run Go tests
make sanity

# Dry-run peribolos locally
mkdir -p ~/.config/peribolos && gh auth token > ~/.config/peribolos/token
make peribolos-dryrun
```

## References
- [Peribolos CLI](https://docs.prow.k8s.io/docs/components/cli-tools/peribolos/)
- [Peribolos source](https://github.com/kubernetes-sigs/prow/tree/main/cmd/peribolos)
- [safe-settings](https://github.com/github/safe-settings)
