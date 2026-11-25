# Maintaining the zero-dot-force GitHub Organization

This document covers operational workflows for managing the zero-dot-force
GitHub organization using two complementary tools: **peribolos** and
**safe-settings**.

## Tool Boundary

| Area | Tool | Config Location |
|------|------|----------------|
| Org membership (admins, members) | peribolos | `org/config.yaml` |
| Team creation, membership, privacy | peribolos | `org/config.yaml` |
| Team-to-repo permission mappings | peribolos | `org/config.yaml` |
| Repo description | peribolos | `org/config.yaml` |
| Repo has_projects | peribolos | `org/config.yaml` |
| Repo default_branch | peribolos | `org/config.yaml` |
| Repo merge strategies | safe-settings | `safe-settings/settings.yml` |
| Repo auto-merge, delete-branch | safe-settings | `safe-settings/settings.yml` |
| Repo has_wiki | safe-settings | `safe-settings/settings.yml` |
| Dependabot alerts and fixes | safe-settings | `safe-settings/settings.yml` |
| Branch protection rules | safe-settings | `safe-settings/suborgs/<suborg>.yml` |
| Rulesets (repo-scoped) | safe-settings | `safe-settings/suborgs/<suborg>.yml` |
| Repo-scoped rulesets, environments | safe-settings | `safe-settings/repos/<repo>.yml` |
| Org name and description | peribolos | `org/config.yaml` |
| Org website | **manual** | GitHub API or UI (peribolos has no field) |
| Repo topics and homepage | safe-settings | `safe-settings/repos/<repo>.yml` |
| Secret scanning, vuln reporting | **manual** | Org security configuration |
| `.github` repo ruleset | **manual** | `rulesets/dot-github-main.json` |

**Why two tools?** Peribolos manages org-level concerns (who is a member,
what teams exist, what permissions teams have). Safe-settings manages
repo-level concerns (how branches are protected, what merge strategies
are allowed, what security features are enabled). This separation follows
the principle of least privilege for their respective GitHub App
permissions.

**Boundary enforcement:** Go tests in `config/boundary_test.go` validate
that neither tool manages fields owned by the other. These tests run on
every PR via CI.

## Common Workflows

### Add or Remove an Org Member

1. Edit `org/config.yaml` -- add/remove the username from the `admins` or
   `members` list (keep sorted alphabetically).
2. If adding, add to the appropriate team(s) as well.
3. Submit a PR. CI validates the config automatically.
4. After merge, peribolos applies the change (push-triggered or daily
   at 05:30 UTC).

### Create a New Team or Change Team Membership

1. Edit `org/config.yaml` -- add/modify the team under the `teams` section.
2. Ensure team members are org members (CI validates this).
3. Ensure admins are listed as `maintainers`, not `members` (CI validates).
4. Submit a PR and merge.

### Add a New Repository

1. Add the repo to `org/config.yaml` under the top-level `repos` section
   with `description`, `has_projects`, and `default_branch` (peribolos-owned
   fields).
2. Add the repo to the appropriate team `repos` mapping.
3. Add the repo to the appropriate suborg file:
   - `safe-settings/suborgs/code-repos.yml` for code repositories
   - `safe-settings/suborgs/non-code-repos.yml` for non-code repositories
4. Add the repo to `restrictedRepos.include` in
   `safe-settings/deployment-settings.yml`. safe-settings only manages
   repos listed there, so a repo missing from it gets none of the org
   defaults, suborg rulesets, or repo overrides.
5. Submit a PR. CI boundary tests validate consistency.
6. After merge, peribolos creates the repo and sets team permissions.
7. Trigger `workflow_dispatch` on the "Safe Settings Sync" workflow to
   apply repo settings.

### Change Branch Protection Rules or Rulesets

1. Edit `safe-settings/suborgs/code-repos.yml` -- modify the ruleset
   under `rulesets`. It is created as a repo-scoped ruleset on each repo
   in the suborg.
2. Do NOT add `rulesets` to `safe-settings/settings.yml`: that creates
   org-level rulesets, which need GitHub Team (zero-dot-force is on
   Free). A boundary test enforces this.
3. Submit a PR and merge.
4. Trigger `workflow_dispatch` to apply.

### Add a Repo-Specific Override

Use repo overrides sparingly. Only create one when a repo needs settings
that differ from its suborg defaults.

1. Create `safe-settings/repos/<repo-name>.yml`.
2. Set only the fields that differ from the suborg/org defaults.
3. Do NOT set peribolos-owned fields (`description`, `has_projects`,
   `default_branch`).
4. Submit a PR. CI boundary tests validate the override.

### Secret Scanning, Push Protection and Vulnerability Reporting

Secret scanning, push protection and private vulnerability reporting come
from the org code security configuration **Zero Dot Force baseline**
(Settings > Code security > Configurations), not from safe-settings. The
pinned safe-settings version skips `security_and_analysis` when it
compares a repo, so it can't keep these settings in place.

- The configuration enables secret scanning, push protection and
  private vulnerability reporting. Every other field is `not_set`, so it
  never touches Dependabot or code scanning.
- Private vulnerability reporting is what makes the **Report a
  vulnerability** button in `.github/SECURITY.md` work.
- It's attached to every repo and is the default for new public repos,
  so a new repo gets it without any extra step.
- `advanced_security` is `enabled` because GitHub requires it for secret
  scanning. On public repos this costs nothing.

To check a repo:

```bash
gh api repos/zero-dot-force/<repo>/code-security-configuration --jq .configuration.name
```

## Override Validator Policies

Override validators in `safe-settings/deployment-settings.yml` enforce
a security floor:

- **Approver count floor**: Suborg or repo configs cannot lower
  `required_approving_review_count` below the org default. Setting it
  higher is allowed.
- **No admin collaborators**: The `admin` permission cannot be granted
  to collaborators via safe-settings. Use peribolos team membership
  with admin role instead.

**Requesting an exception:** If a legitimate use case requires bypassing
a validator, discuss with org admins. Exceptions require modifying the
validator script in `deployment-settings.yml` via a reviewed PR.

## Local Validation

### Prerequisites

- Go (version in `go.mod`)
- `yamllint` (for YAML validation)

### Commands

```bash
# Validate all YAML (peribolos + safe-settings)
make lint

# Run all Go tests (peribolos + boundary)
make test-unit

# Validate only safe-settings YAML
make safe-settings-validate

# Full validation: format, vet, lint, tests, diff check
make sanity
```

## Applying Safe-settings Changes

safe-settings reads its config from the `.github` repo's default branch
via the GitHub API. Config changes must be **merged to main** before
safe-settings can apply them.

### Testing sequence

1. **Local validation** (before PR):
   ```bash
   make test-unit              # boundary tests
   make safe-settings-validate # YAML syntax
   ```

2. **Submit PR** -- CI runs boundary tests and YAML validation.

3. **Merge PR** -- config lands on main.

4. **Dry-run against a single repo** -- go to Actions > "Safe Settings
   Sync" > "Run workflow":
   - Set `dry-run` to `true`
   - Set `repos` to a single repo (e.g., `snake-eyes`)
   - Review the workflow output to see what would change

5. **Apply to a single repo** -- same workflow:
   - Set `dry-run` to `false`
   - Set `repos` to the same repo

6. **Apply to all repos** -- same workflow:
   - Set `dry-run` to `false`
   - Leave `repos` empty (applies to all managed repos)

### Rollback

If safe-settings applies incorrect settings:
1. `git revert` the config change and push to main
2. Trigger `workflow_dispatch` with `dry-run=false` -- safe-settings
   reverts to the previous config state
3. Or fix settings manually via the GitHub UI (safe-settings will
   re-apply them on the next sync)

## Triggering Manual Sync

### Peribolos

Go to Actions > "Peribolos: Apply" > "Run workflow". Set `dry-run` to
`true` for a preview, or `false` to apply.

### Safe-settings

Go to Actions > "Safe Settings Sync" > "Run workflow":
- **dry-run**: `true` to preview, `false` to apply (defaults to `true`)
- **repos**: comma-separated list of repos to target (e.g.,
  `snake-eyes,reading-stone`). Leave empty to apply to all managed
  repos.

### Future automation

After initial validation, the workflow can be extended with:
- `push` trigger on `safe-settings/**` path changes to main
- `schedule` trigger (daily at 06:00 UTC) for drift correction

These triggers are intentionally disabled during the initial rollout to
ensure full manual control.

## Troubleshooting

### Settings not applied after merge

1. Trigger `workflow_dispatch` manually -- safe-settings only runs on
   manual dispatch during initial rollout (no push/schedule triggers).
2. Check the "Safe Settings Sync" workflow run in the Actions tab.
3. Look for errors in the workflow logs (credential expiry, API errors).

### Boundary test failures

Boundary tests fail when:
- A repo in a suborg file does not exist in `org/config.yaml` -- add it
  to peribolos first.
- A repo appears in multiple suborg files -- each repo belongs to exactly
  one suborg.
- A safe-settings config sets `description`, `has_projects`, or
  `default_branch` -- these are peribolos-owned fields.
- A suborg repo list does not match the corresponding ruleset
  `repository_name.include` -- update both files together.

### safe-settings sync errors

Common causes:
- **Credential expiry**: The GitHub App private key may need rotation.
  Update the `SAFE_SETTINGS_PRIVATE_KEY` secret.
- **API rate limits**: The sync may fail if it hits GitHub API rate
  limits. Wait and re-trigger.
- **Invalid YAML**: The workflow validates YAML before applying. Check
  the yamllint output in the workflow logs.
- **safe-settings version issue**: If safe-settings behavior changes,
  check the pinned version in the workflow file.

### Known upstream workarounds

The workflow includes a patched `full-sync.js` to work around a bug
in safe-settings where `handleResults` crashes in full-sync mode
because `payload.check_suite` is undefined outside the webhook flow
(the sync itself completes; only the Check Run reporting fails). The
patch replaces `handleResults` to log the results and fail the step
when any plugin returns an error, so a failed sync never reports
success.

- **Upstream issue**:
  [github-community-projects/safe-settings#818](https://github.com/github-community-projects/safe-settings/issues/818)
- **Upstream fix PR**:
  [github-community-projects/safe-settings#1018](https://github.com/github-community-projects/safe-settings/pull/1018)
- **Search tag**: `TODO(safe-settings-818)` in the workflow file

Once upstream PR #1018 is merged and released, update the pinned
version and revert the patched script back to `npm run full-sync`.

## Excluded Repos

The following repos are excluded from safe-settings management:

- `.github` -- the admin repo (avoids circular dependency). Its
  ruleset is applied by hand from `rulesets/dot-github-main.json`.

### The `.github` ruleset

A push to `main` here changes org membership, teams and every repo's
settings, so `main` is protected by the ruleset
**main: reviewed changes only**:

- one approving review, from a code owner in `CODEOWNERS`, and a new
  approval after any later push
- the `Lint`, `Test` and `Verify peribolos` checks must pass
- no force pushes, and `main` can't be deleted

Org admins can bypass it, but only when merging a pull request. Direct
pushes stay blocked for everyone. Each bypass is logged under the
ruleset's **Insights** tab.

`rulesets/dot-github-main.json` is the source of truth. After changing
it, apply it with:

```bash
id=$(gh api repos/zero-dot-force/.github/rulesets \
  --jq '.[] | select(.name == "main: reviewed changes only") | .id')
gh api -X PUT "repos/zero-dot-force/.github/rulesets/$id" \
  --input rulesets/dot-github-main.json
```

These are listed in `safe-settings/deployment-settings.yml` under
`restrictedRepos` and/or excluded from suborg files.
