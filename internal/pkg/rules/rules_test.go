package rules

/*
integration test for rules. This separate pkg needed to not create an import cycle
between config and rules
*/

import (
	"reflect"
	"sort"
	"testing"

	_ "github.com/tumblr/docker-registry-pruner/internal/pkg/testing"
	"github.com/tumblr/docker-registry-pruner/pkg/config"
	"github.com/tumblr/docker-registry-pruner/pkg/rules"
)

var (
	rulesDir   = "test/fixtures/rules"
	testImages = map[string][]string{
		"tumblr/bb8": []string{"v0.6.0-480-g5d09186", "v0.6.0-486-g77397a0", "v0.6.0-413-g463a787", "latest", "v4.2.0", "v4.2.1", "some-ignored-tag", "anothertag", "0.1.2+notignored"},
		"gar/nix":    []string{},
		"foo/bar":    []string{"1.2.3", "abf273", "henlo"},
		"image/x":    []string{"v0.1.1+x", "v0.6.9+x", "v4.2.1+x", "0.0.1+x", "0.0.2+x"},
		"image/y":    []string{"v0.1.0+y", "v0.69.420+y", "v4.2.0+y", "0.0.1+y", "0.0.2+y"},
		"tumblr/redpop": []string{
			"abcdef123", "v1.2.3", "v1.2.4", "v1.2.5", "v2.0+hello", "pr-123", "pr-124", "pr-2345",
		},
	}
)

func TestLoadRules(t *testing.T) {
	tests := map[string]int{
		"bb8-ignore-some.yaml":   2,
		"bb8-match-version.yaml": 2,
		"bb8-match-all.yaml":     2,
		"redpop-pr.yaml":         1,
		"bb8-multiple.yaml":      3,
		"multiple-repos.yaml":    3,
	}
	for f, nExpected := range tests {
		cfg, err := config.LoadFromFile(rulesDir + "/" + f)
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

func TestMatching(t *testing.T) {
	fixturesExpected := map[string][]string{
		"bb8-match-all.yaml":     []string{"v0.6.0-480-g5d09186", "v0.6.0-486-g77397a0", "v0.6.0-413-g463a787", "v4.2.0", "v4.2.1", "some-ignored-tag", "0.1.2+notignored", "anothertag"},
		"bb8-match-version.yaml": []string{"v0.6.0-480-g5d09186", "v0.6.0-486-g77397a0", "v0.6.0-413-g463a787", "v4.2.0", "v4.2.1"},
		"bb8-ignore-some.yaml":   []string{"0.1.2+notignored", "anothertag"},
	}
	repo := "tumblr/bb8"
	for f, expectedTags := range fixturesExpected {
		fixture := rulesDir + "/" + f
		t.Logf("loading rules %s for %s", fixture, repo)
		tags := testImages[repo]

		cfg, err := config.LoadFromFile(fixture)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		t.Logf("Loaded %d rules\n", len(cfg.Rules))
		selectors := rules.RulesToSelectors(cfg.Rules)

		foundTags := []string{}
		for _, tag := range tags {
			if rules.MatchAny(selectors, repo, tag) {
				foundTags = append(foundTags, tag)
			}
		}

		sort.Strings(foundTags)
		sort.Strings(expectedTags)
		if !reflect.DeepEqual(expectedTags, foundTags) {
			t.Errorf("%s: expected matching tags to be %v but got %v", fixture, expectedTags, foundTags)
			t.Fail()
		}
	}
}

func TestFilterRepoTags(t *testing.T) {
	repoTags := testImages
	fixturesRulesToExpected := map[string]map[string][]string{
		"bb8-ignore-some.yaml": map[string][]string{
			"tumblr/bb8": []string{"0.1.2+notignored", "anothertag"},
		},
		"bb8-match-all.yaml": map[string][]string{
			"tumblr/bb8": []string{"0.1.2+notignored", "anothertag", "some-ignored-tag", "v0.6.0-413-g463a787", "v0.6.0-480-g5d09186", "v0.6.0-486-g77397a0", "v4.2.0", "v4.2.1"},
		},
		"bb8-match-version.yaml": map[string][]string{
			"tumblr/bb8": []string{"v0.6.0-413-g463a787", "v0.6.0-480-g5d09186", "v0.6.0-486-g77397a0", "v4.2.0", "v4.2.1"},
		},
		"redpop-pr.yaml": map[string][]string{
			"tumblr/redpop": []string{"pr-123", "pr-124", "pr-2345"},
		},
		"bb8-tagselectors.yaml": map[string][]string{
			"tumblr/bb8": []string{"v0.6.0-413-g463a787", "v0.6.0-480-g5d09186", "v0.6.0-486-g77397a0", "v4.2.0", "v4.2.1"},
		},
		"multiple-repos.yaml": map[string][]string{
			"tumblr/bb8":    []string{"v0.6.0-413-g463a787", "v0.6.0-480-g5d09186", "v0.6.0-486-g77397a0", "v4.2.0", "v4.2.1"},
			"tumblr/redpop": []string{"pr-123", "pr-124", "pr-2345"},
		},
		"multiple-repo-versions.yaml": map[string][]string{
			"image/x": []string{"0.0.1+x", "0.0.2+x", "v0.1.1+x", "v0.6.9+x", "v4.2.1+x"},
			"image/y": []string{"0.0.1+y", "0.0.2+y", "v0.1.0+y", "v0.69.420+y", "v4.2.0+y"},
		},
	}
	for fixtureFile, expected := range fixturesRulesToExpected {
		fixture := rulesDir + "/" + fixtureFile
		t.Logf("Applying filters from fixture %s...", fixture)

		cfg, err := config.LoadFromFile(fixture)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		selectors := rules.RulesToSelectors(cfg.Rules)

		actualRepoTags := rules.FilterRepoTags(repoTags, selectors)
		if !reflect.DeepEqual(expected, actualRepoTags) {
			t.Errorf("%s: expected matching tags to be %v but got %v", fixture, expected, actualRepoTags)
			t.Fail()
		}
	}
}
