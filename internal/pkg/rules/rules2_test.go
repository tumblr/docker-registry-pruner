package rules

import (
	"reflect"
	"sort"
	"testing"
	"time"

	_ "github.com/tumblr/docker-registry-pruner/internal/pkg/testing"
	"github.com/tumblr/docker-registry-pruner/pkg/config"
	"github.com/tumblr/docker-registry-pruner/pkg/registry"
	"github.com/tumblr/docker-registry-pruner/pkg/rules"
)

// helper function to make a test fixture manifest
func mkmanifest(r, t string, daysOld int64, labels map[string]string) *registry.Manifest {
	if labels == nil {
		labels = map[string]string{}
	}
	return must(registry.NewManifest(r, t, tNow.Add(time.Duration(-daysOld*24)*time.Hour), labels))
}

func must(m *registry.Manifest, err error) *registry.Manifest {
	if err != nil {
		panic(err)
	}
	return m
}

var (
	tNow = time.Now()
)

type testpayload struct {
	rulesFile    string
	input        []*registry.Manifest
	keepImages   []string
	deleteImages []string
}

func TestApplyRules(t *testing.T) {

	tc, err := loadTestConfig("test/fixtures/manifest_tests/apply-rules.yaml")
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

		keep, delete := rules.ApplyRules(cfg.Rules, tc.Manifests)
		keep_tags := manifestsAsImageMap(keep)
		delete_tags := manifestsAsImageMap(delete)
		if test.Config == "test/fixtures/rules/labels-devel-3-versions.yaml" {
			t.Logf("%s: expected: %+v", test.Config, test)
			t.Logf("%s: kept: %+v", test.Config, keep_tags)
			t.Logf("%s: deleted: %+v", test.Config, delete_tags)
		}

		if !reflect.DeepEqual(test.Expected.Keep, keep_tags) {
			t.Errorf("%s: expected keep images to be %v but was actually %v", test.Config, test.Expected.Keep, keep_tags)
			t.FailNow()
		}
		if !reflect.DeepEqual(test.Expected.Delete, delete_tags) {
			t.Errorf("%s: expected delete images tags to be %v but was actually %v", test.Config, test.Expected.Delete, delete_tags)
			t.FailNow()
		}
	}
}

// turn a list of Manifest into a map of repo->list of tags
func manifestsAsImageMap(ms []*registry.Manifest) map[string][]string {
	res := map[string][]string{}
	for _, m := range ms {
		if _, ok := res[m.Name]; !ok {
			res[m.Name] = []string{}
		}
		res[m.Name] = append(res[m.Name], m.Tag)
		sort.Strings(res[m.Name])
	}
	return res
}
