package config

import (
	"testing"

	_ "github.com/tumblr/docker-registry-pruner/internal/pkg/testing"
	"github.com/tumblr/docker-registry-pruner/pkg/rules"
)

type errorFixture struct {
	expected error
	file     string
}

var (
	rulesDir         = "test/fixtures/rules"
	fixtureDirectory = "test/fixtures/config"
	errorTests       = []errorFixture{
		{
			file:     "invalid-missing-registry.yaml",
			expected: ErrMissingRegistry,
		},
		{
			file:     "invalid-rule-missing-repos-and-labels.yaml",
			expected: rules.ErrMissingReposOrLabels,
		},
		{
			file:     "invalid-rule-missing-action.yaml",
			expected: rules.ErrActionMustBeSpecified,
		},
		{
			file:     "invalid-rule-no-rules.yaml",
			expected: ErrNoRulesLoaded,
		},
		{
			file:     "invalid-rule-duplicate-action-days-latest.yaml",
			expected: rules.ErrMultipleActionDaysLatest,
		},
		{
			file:     "invalid-rule-duplicate-action-days-versions.yaml",
			expected: rules.ErrMultipleActionVersionsDays,
		},
		{
			file:     "invalid-rule-duplicate-action-versions-latest.yaml",
			expected: rules.ErrMultipleActionLatestVersions,
		},
	}
)

func TestLoadInvalidConfigs(t *testing.T) {
	for _, test := range errorTests {
		f := fixtureDirectory + "/" + test.file
		t.Logf("loading config from %s", f)
		_, err := LoadFromFile(f)
		if err != test.expected {
			t.Errorf("Expected loading %s to produce error %v, but got %v", test.file, test.expected, err)
			t.Fail()
		}
	}
}

func TestLoadRules(t *testing.T) {
	tests := map[string]int{
		"fleeble-ignore-some.yaml":   2,
		"fleeble-match-version.yaml": 2,
		"fleeble-match-all.yaml":     2,
		"plumbus-pr.yaml":            1,
		"fleeble-multiple.yaml":      3,
		"multiple-repos.yaml":        3,
	}
	for f, nExpected := range tests {
		cfg, err := LoadFromFile(rulesDir + "/" + f)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		if len(cfg.Rules) != nExpected {
			t.Errorf("%s: expected %d rules loaded but found %d", f, nExpected, len(cfg.Rules))
			t.Fail()
		}
		t.Logf("Loaded %d rules\n", len(cfg.Rules))
	}
}
