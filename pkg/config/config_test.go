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
	fixtureDirectory = "test/fixtures/config"
	errorTests       = []errorFixture{
		{
			file:     "invalid-missing-registry.yaml",
			expected: ErrMissingRegistry,
		},
		{
			file:     "invalid-rule-missing-repos.yaml",
			expected: rules.ErrMissingRepos,
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
