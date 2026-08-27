// SPDX-License-Identifier: Apache-2.0

// Package config provides validation tests for the zero-dot-force org configuration.
//
// boundary_test.go validates cross-tool consistency between org/config.yaml
// and safe-settings config. It ensures:
//   - All repos in suborg files exist in org/config.yaml
//   - No repo appears in multiple suborg files
//   - Safe-settings config does not set fields owned by peribolos
//   - Suborg repo lists match ruleset repository_name conditions
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
)

// safeSettingsDir is the path to the safe-settings config directory.
var safeSettingsDir = "../safe-settings"

// peribolosOwnedFields are repo-level fields that peribolos manages.
// Safe-settings config must NOT set these fields.
var peribolosOwnedFields = []string{
	"description",
	"has_projects",
	"default_branch",
}

// safeSettingsOwnedFields are repo-level fields that safe-settings manages.
// Peribolos config must NOT set these fields.
var safeSettingsOwnedFields = []string{
	"has_wiki",
	"has_issues", // reserved; not actively managed in initial deployment

	"allow_auto_merge",
	"delete_branch_on_merge",
	"allow_squash_merge",
	"allow_merge_commit",
	"allow_rebase_merge",
	"allow_update_branch",
}

// suborg represents a parsed suborg configuration file.
type suborg struct {
	SuborgRepos []string `json:"suborgrepos"`
}

// settingsFile represents a parsed safe-settings settings.yml.
type settingsFile struct {
	Repository map[string]interface{} `json:"repository"`
	Rulesets   []ruleset              `json:"rulesets"`
}

// ruleset represents a parsed safe-settings ruleset.
type ruleset struct {
	Name       string            `json:"name"`
	Conditions rulesetConditions `json:"conditions"`
}

// rulesetConditions represents the conditions block of a ruleset.
type rulesetConditions struct {
	RepositoryName *repositoryNameCondition `json:"repository_name"`
}

// repositoryNameCondition represents the repository_name condition.
type repositoryNameCondition struct {
	Include []string `json:"include"`
}

// repoOverride represents a parsed repo-level override file.
type repoOverride struct {
	Repository map[string]interface{} `json:"repository"`
}

// loadSuborgs parses all suborg YAML files from the suborgs directory.
func loadSuborgs(dir string) (map[string][]string, error) {
	suborgsDir := filepath.Join(dir, "suborgs")
	entries, err := os.ReadDir(suborgsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading suborgs directory: %w", err)
	}

	result := make(map[string][]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(suborgsDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading suborg file %s: %w", entry.Name(), err)
		}

		var s suborg
		if err := yaml.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("parsing suborg file %s: %w", entry.Name(), err)
		}

		result[entry.Name()] = s.SuborgRepos
	}

	return result, nil
}

// loadSettings parses the main settings.yml file.
func loadSettings(dir string) (*settingsFile, error) {
	data, err := os.ReadFile(filepath.Join(dir, "settings.yml"))
	if err != nil {
		return nil, fmt.Errorf("reading settings.yml: %w", err)
	}

	var s settingsFile
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing settings.yml: %w", err)
	}

	return &s, nil
}

// loadRepoOverrides parses all repo-level override YAML files.
func loadRepoOverrides(dir string) (map[string]*repoOverride, error) {
	reposDir := filepath.Join(dir, "repos")
	entries, err := os.ReadDir(reposDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading repos directory: %w", err)
	}

	result := make(map[string]*repoOverride)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(reposDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading repo override file %s: %w", entry.Name(), err)
		}

		var r repoOverride
		if err := yaml.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("parsing repo override file %s: %w", entry.Name(), err)
		}

		result[entry.Name()] = &r
	}

	return result, nil
}

// peribolosRepoNames extracts the repo names from the parsed peribolos
// config. The peribolos YAML structure is orgs.<orgname>.repos.<reponame>.
func peribolosRepoNames() map[string]bool {
	repos := make(map[string]bool)
	for _, orgCfg := range cfg.Orgs {
		for repoName := range orgCfg.Repos {
			repos[repoName] = true
		}
	}
	return repos
}

// peribolosRepoFields returns the set of fields configured per repo in
// org/config.yaml. It re-parses the raw YAML to detect field names
// without relying on Go struct field mappings.
func peribolosRepoFields(configFile string) (map[string]map[string]bool, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("reading peribolos config: %w", err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing peribolos config: %w", err)
	}

	result := make(map[string]map[string]bool)

	orgs, ok := raw["orgs"].(map[string]interface{})
	if !ok {
		return result, nil
	}

	for _, orgData := range orgs {
		org, ok := orgData.(map[string]interface{})
		if !ok {
			continue
		}

		repos, ok := org["repos"].(map[string]interface{})
		if !ok {
			continue
		}

		for repoName, repoData := range repos {
			repoFields, ok := repoData.(map[string]interface{})
			if !ok {
				continue
			}

			fields := make(map[string]bool)
			for fieldName := range repoFields {
				fields[fieldName] = true
			}
			result[repoName] = fields
		}
	}

	return result, nil
}

func TestBoundary_SuborgReposExistInPeribolos(t *testing.T) {
	if _, err := os.Stat(safeSettingsDir); os.IsNotExist(err) {
		t.Skip("safe-settings directory not found, skipping boundary tests")
	}

	peribolosRepos := peribolosRepoNames()
	suborgs, err := loadSuborgs(safeSettingsDir)
	if err != nil {
		t.Fatalf("failed to load suborgs: %v", err)
	}

	for fileName, repos := range suborgs {
		for _, repo := range repos {
			if !peribolosRepos[repo] {
				t.Errorf("suborg file %s references repo %q which does not exist in org/config.yaml", fileName, repo)
			}
		}
	}
}

func TestBoundary_NoDuplicateSuborgMembership(t *testing.T) {
	if _, err := os.Stat(safeSettingsDir); os.IsNotExist(err) {
		t.Skip("safe-settings directory not found, skipping boundary tests")
	}

	suborgs, err := loadSuborgs(safeSettingsDir)
	if err != nil {
		t.Fatalf("failed to load suborgs: %v", err)
	}

	repoToSuborg := make(map[string]string)
	for fileName, repos := range suborgs {
		for _, repo := range repos {
			if existing, ok := repoToSuborg[repo]; ok {
				t.Errorf("repo %q appears in both %s and %s", repo, existing, fileName)
			}
			repoToSuborg[repo] = fileName
		}
	}
}

func TestBoundary_SafeSettingsNoPeriblosFields(t *testing.T) {
	if _, err := os.Stat(safeSettingsDir); os.IsNotExist(err) {
		t.Skip("safe-settings directory not found, skipping boundary tests")
	}

	// Check settings.yml
	settings, err := loadSettings(safeSettingsDir)
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}

	if settings.Repository != nil {
		for _, field := range peribolosOwnedFields {
			if _, exists := settings.Repository[field]; exists {
				t.Errorf("settings.yml sets peribolos-owned field %q under repository", field)
			}
		}
	}

	// Check suborg files (re-parse as raw YAML for field detection)
	suborgsDir := filepath.Join(safeSettingsDir, "suborgs")
	entries, err := os.ReadDir(suborgsDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to read suborgs directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(suborgsDir, entry.Name()))
		if err != nil {
			t.Fatalf("failed to read suborg file %s: %v", entry.Name(), err)
		}

		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			t.Fatalf("failed to parse suborg file %s: %v", entry.Name(), err)
		}

		if repo, ok := raw["repository"].(map[string]interface{}); ok {
			for _, field := range peribolosOwnedFields {
				if _, exists := repo[field]; exists {
					t.Errorf("suborg file %s sets peribolos-owned field %q under repository", entry.Name(), field)
				}
			}
		}
	}

	// Check repo-level override files
	overrides, err := loadRepoOverrides(safeSettingsDir)
	if err != nil {
		t.Fatalf("failed to load repo overrides: %v", err)
	}

	for fileName, override := range overrides {
		if override.Repository != nil {
			for _, field := range peribolosOwnedFields {
				if _, exists := override.Repository[field]; exists {
					t.Errorf("repo override file %s sets peribolos-owned field %q under repository", fileName, field)
				}
			}
		}
	}
}

func TestBoundary_PeribolosNoSafeSettingsFields(t *testing.T) {
	repoFields, err := peribolosRepoFields(*configPath)
	if err != nil {
		t.Fatalf("failed to parse peribolos repo fields: %v", err)
	}

	for repoName, fields := range repoFields {
		for _, forbidden := range safeSettingsOwnedFields {
			if fields[forbidden] {
				t.Errorf("org/config.yaml repo %q sets safe-settings-owned field %q", repoName, forbidden)
			}
		}
	}
}

func TestBoundary_SuborgReposMatchRulesetConditions(t *testing.T) {
	if _, err := os.Stat(safeSettingsDir); os.IsNotExist(err) {
		t.Skip("safe-settings directory not found, skipping boundary tests")
	}

	suborgs, err := loadSuborgs(safeSettingsDir)
	if err != nil {
		t.Fatalf("failed to load suborgs: %v", err)
	}

	settings, err := loadSettings(safeSettingsDir)
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}

	// Build a map of suborg name (without extension) to repo set
	suborgRepoSets := make(map[string]map[string]bool)
	for fileName, repos := range suborgs {
		name := strings.TrimSuffix(fileName, ".yml")
		repoSet := make(map[string]bool)
		for _, r := range repos {
			repoSet[r] = true
		}
		suborgRepoSets[name] = repoSet
	}

	// For each ruleset, check if its repository_name.include matches a
	// suborg. The naming convention is "safe-settings: <group>" where
	// <group> corresponds to a suborg file name (with hyphens replaced
	// by spaces). E.g., "safe-settings: code repos" matches "code-repos.yml".
	for _, rs := range settings.Rulesets {
		if rs.Conditions.RepositoryName == nil {
			continue
		}

		rulesetRepos := make(map[string]bool)
		for _, r := range rs.Conditions.RepositoryName.Include {
			rulesetRepos[r] = true
		}

		// Normalize the ruleset name for matching: strip the
		// "safe-settings: " prefix, then replace spaces with hyphens
		// to match suborg file names (e.g., "code repos" -> "code-repos").
		normalized := strings.TrimPrefix(rs.Name, "safe-settings: ")
		normalized = strings.ReplaceAll(normalized, " ", "-")

		for suborgName, suborgRepos := range suborgRepoSets {
			if normalized != suborgName {
				continue
			}

			// Check repos in suborg but not in ruleset
			for repo := range suborgRepos {
				if !rulesetRepos[repo] {
					t.Errorf("repo %q is in suborg %s but missing from ruleset %q repository_name.include",
						repo, suborgName+".yml", rs.Name)
				}
			}

			// Check repos in ruleset but not in suborg
			for repo := range rulesetRepos {
				if !suborgRepos[repo] {
					t.Errorf("repo %q is in ruleset %q repository_name.include but missing from suborg %s",
						repo, rs.Name, suborgName+".yml")
				}
			}
		}
	}
}
