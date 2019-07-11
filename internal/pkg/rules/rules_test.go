package rules

/*
integration test for rules. This separate pkg needed to not create an import cycle
between config and rules
*/

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"reflect"
	"sort"
	"testing"

	_ "github.com/tumblr/docker-registry-pruner/internal/pkg/testing"
	"github.com/tumblr/docker-registry-pruner/pkg/config"
	"github.com/tumblr/docker-registry-pruner/pkg/registry"
	"github.com/tumblr/docker-registry-pruner/pkg/rules"
)

func manifestObjectsToManifests(objs []*manifestObject) []*registry.Manifest {
	ms := []*registry.Manifest{}
	for _, o := range objs {
		m := mkmanifest(o.Name, o.Tag, o.DaysOld, o.Labels)
		ms = append(ms, m)
	}
	return ms
}

// testConfig is a configuration that defines a set of test. It is comprised of:
// * SourceManifests: all Manifests that will be parsed into a registry.Manifest via mkmanifest. These are source material for the test suite
// * Tests: List of `testCase`
type testConfig struct {
	SourceFile      string
	Manifests       []*registry.Manifest
	SourceManifests []*manifestObject `yaml:"source_manifests"`
	Tests           []testCase        `yaml:"tests"`
}

// manifestObject will be parsed from test configs, and then pumped into mkmanifest()
// to turn it into a registry.Manifest.
type manifestObject struct {
	Name    string
	Tag     string
	DaysOld int64             `yaml:"days_old"`
	Labels  map[string]string `yaml:"labels"`
}

// testCase is a struct to define a specific test case. It is comprised of:
// * Config: the yaml config that contains the Rule sets
// * Expected: The map[repo][]tags that the rest should produce from the testConfig.Manifests as input
type testCase struct {
	Config   string `yaml:"config"`
	Expected struct {
		Keep   map[string][]string `yaml:"keep"`
		Delete map[string][]string `yaml:"delete"`
	} `yaml:"expected"`
}

func loadTestConfig(cfg string) (*testConfig, error) {
	d, err := ioutil.ReadFile(cfg)
	if err != nil {
		return nil, err
	}

	tc := testConfig{}
	err = yaml.Unmarshal(d, &tc)
	if err != nil {
		return nil, err
	}
	tc.SourceFile = cfg
	tc.Manifests = manifestObjectsToManifests(tc.SourceManifests)

	return &tc, nil
}

func TestMatching(t *testing.T) {
	tc, err := loadTestConfig("test/fixtures/manifest_tests/manifest_matching_1.yaml")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, test := range tc.Tests {
		// sort any expected tag sets
		for _, tags := range test.Expected.Keep {
			sort.Strings(tags)
		}
		for _, tags := range test.Expected.Delete {
			sort.Strings(tags)
		}
		t.Logf("%s: loading rules from %s", tc.SourceFile, test.Config)

		cfg, err := config.LoadFromFile(test.Config)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		t.Logf("%s: Loaded %d rules from %s (%d manifests)", tc.SourceFile, len(cfg.Rules), test.Config, len(tc.Manifests))
		selectors := rules.RulesToSelectors(cfg.Rules)

		matchedManifests := map[string][]string{}
		for _, manifest := range tc.Manifests {
			if rules.MatchAny(selectors, manifest) {
				if matchedManifests[manifest.Name] == nil {
					matchedManifests[manifest.Name] = []string{}
				}
				matchedManifests[manifest.Name] = append(matchedManifests[manifest.Name], manifest.Tag)
				sort.Strings(matchedManifests[manifest.Name])
			}
		}

		if !reflect.DeepEqual(test.Expected.Keep, matchedManifests) {
			t.Errorf("%s: (rules %s) expected matching tags to be %v but got %v", tc.SourceFile, test.Config, test.Expected.Keep, matchedManifests)
			t.FailNow()
		}
	}
}

func TestFilterRepoTags(t *testing.T) {
	tc, err := loadTestConfig("test/fixtures/manifest_tests/filter_repo_tags.yaml")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, test := range tc.Tests {
		// sort any expected tag sets
		for _, tags := range test.Expected.Keep {
			sort.Strings(tags)
		}
		for _, tags := range test.Expected.Keep {
			sort.Strings(tags)
		}
		t.Logf("%s: loading rules from %s", tc.SourceFile, test.Config)

		cfg, err := config.LoadFromFile(test.Config)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		selectors := rules.RulesToSelectors(cfg.Rules)

		actualManifests := rules.FilterManifests(tc.Manifests, selectors)
		// construct a map[string]map[string][]string from actualManifests to aid in comparison
		actualManifestsTags := map[string][]string{}
		for repo, ms := range actualManifests {
			actualManifestsTags[repo] = []string{}
			for _, m := range ms {
				actualManifestsTags[repo] = append(actualManifestsTags[repo], m.Tag)
			}
			sort.Strings(actualManifestsTags[repo])
		}

		if !reflect.DeepEqual(test.Expected.Keep, actualManifestsTags) {
			t.Errorf("%s: (rules %s) expected matching tags to be:\n%v\nbut got:\n%v", tc.SourceFile, test.Config, test.Expected.Keep, actualManifestsTags)
			t.FailNow()
		}
	}
}
