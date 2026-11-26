// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
)

type safeSettingsLabelFile struct {
	Labels *safeSettingsLabels `json:"labels"`
}

type safeSettingsLabels struct {
	Include []safeSettingsLabel `json:"include"`
	Exclude []labelExclusion    `json:"exclude"`
}

type safeSettingsLabel struct {
	Name  string      `json:"name"`
	Color interface{} `json:"color"`
}

type labelExclusion struct {
	Name string `json:"name"`
}

func TestSafeSettingsLabels_UndeclaredLabelsArePreserved(t *testing.T) {
	for fileName, labels := range loadSafeSettingsLabelDeclarations(t) {
		if len(labels.Include) == 0 {
			t.Errorf("%s labels.include is empty", fileName)
		}
		if len(labels.Exclude) != 1 || labels.Exclude[0].Name != ".*" {
			t.Errorf("%s labels.exclude = %#v, want one catch-all exclusion that preserves undeclared labels", fileName, labels.Exclude)
		}
	}
}

func TestSafeSettingsLabels_CIColorIsString(t *testing.T) {
	declarations := loadSafeSettingsLabelDeclarations(t)
	settingsLabels := declarations["settings.yml"]
	for _, label := range settingsLabels.Include {
		if label.Name != "ci" {
			continue
		}

		color, ok := label.Color.(string)
		if !ok {
			t.Fatalf("ci label color has type %T, want string", label.Color)
		}
		if color != "5319e7" {
			t.Fatalf("ci label color = %q, want %q", color, "5319e7")
		}
		return
	}

	t.Fatal("settings.yml has no ci label in labels.include")
}

func loadSafeSettingsLabelDeclarations(t *testing.T) map[string]safeSettingsLabels {
	t.Helper()

	paths := []string{filepath.Join(safeSettingsDir, "settings.yml")}
	repoEntries, err := os.ReadDir(filepath.Join(safeSettingsDir, "repos"))
	if err != nil {
		t.Fatalf("failed to read safe-settings repo overrides: %v", err)
	}
	for _, entry := range repoEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yml") {
			paths = append(paths, filepath.Join(safeSettingsDir, "repos", entry.Name()))
		}
	}

	declarations := make(map[string]safeSettingsLabels)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}

		var config safeSettingsLabelFile
		if err := yaml.Unmarshal(data, &config); err != nil {
			t.Fatalf("failed to parse %s: %v", path, err)
		}
		if config.Labels == nil {
			continue
		}

		fileName, err := filepath.Rel(safeSettingsDir, path)
		if err != nil {
			t.Fatalf("failed to identify safe-settings file %s: %v", path, err)
		}
		declarations[filepath.ToSlash(fileName)] = *config.Labels
	}

	if len(declarations) == 0 {
		t.Fatalf("no label declarations found under %s", safeSettingsDir)
	}
	return declarations
}
