package registry

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	_ "github.com/tumblr/docker-registry-pruner/internal/pkg/testing"
	"github.com/tumblr/docker-registry-pruner/pkg/config"
)

// helper function to make a test fixture manifest
func mkmanifest(r, t string, daysOld int64) *Manifest {
	return must(NewManifestWithLastModified(r, t, tNow.Add(time.Duration(-daysOld*24)*time.Hour)))
}

var (
	rulesDir = "test/fixtures/rules"

	tNow      = time.Now()
	manifests = []*Manifest{
		mkmanifest("tumblr/redpop", "v1.2.3", 4),
		mkmanifest("tumblr/redpop", "v1.0.3+metadata", 69),
		mkmanifest("tumblr/redpop", "pr-69420+13d", 13),
		mkmanifest("tumblr/redpop", "pr-69420+14d", 14),
		mkmanifest("tumblr/redpop", "master-v1.2.3-69", 14),
		mkmanifest("tumblr/redpop", "master-2019", 14),
		mkmanifest("tumblr/redpop", "pr-420", 1),
		mkmanifest("tumblr/redpop", "pr-69", 5),
		mkmanifest("tumblr/redpop", "pr-69421+15d", 15),
		mkmanifest("tumblr/redpop", "pr-69419+16d", 16),
		mkmanifest("image/latest", "latest", 69420),
		mkmanifest("tumblr/bb8", "latest", 0),
		mkmanifest("tumblr/bb8", "garbage", 5),
		mkmanifest("tumblr/bb8", "v0.4.2-259-something", 5),
		mkmanifest("tumblr/bb8", "v0.5.0-260", 4),
		mkmanifest("tumblr/bb8", "v0.5.1-260", 4),
		mkmanifest("tumblr/bb8", "v0.5.23+test", 3),
		mkmanifest("tumblr/bb8", "v0.5.2", 3),
		mkmanifest("tumblr/bb8", "some-ignored-tag", 69),
		mkmanifest("tumblr/bb8", "oldtag-1", 69),
		mkmanifest("tumblr/bb8", "oldtag-2", 70),
		mkmanifest("tumblr/bb8", "v0.6.1-261-gbb41394", 0),
		mkmanifest("tumblr/bb8", "v0.5.3-nice", 1),
		mkmanifest("tumblr/bb8", "v0.69-6969", 1),
		mkmanifest("tumblr/bb8", "v0.5.5-420", 1),
		mkmanifest("tumblr/bb8", "v0.6.1-262", 0),
		mkmanifest("tumblr/bb8", "branch-v1.2.3-69", 14),
		mkmanifest("tumblr/bb8", "v0.69.1-262", 0),
		mkmanifest("tumblr/bb8", "abc123f", 0),

		mkmanifest("image/x", "v0.1.1+x", 0),
		mkmanifest("image/x", "v0.6.9+x", 0),
		mkmanifest("image/x", "v4.2.1+x", 0),
		mkmanifest("image/x", "0.0.1+x", 0),
		mkmanifest("image/x", "0.0.2+x", 0),
		mkmanifest("image/y", "v0.1.0+y", 0),
		mkmanifest("image/y", "v0.69.420+y", 0),
		mkmanifest("image/y", "v4.2.0+y", 0),
		mkmanifest("image/y", "0.0.1+y", 0),
		mkmanifest("image/y", "0.0.2+y", 0),
	}
	tests = []testpayload{
		{
			rulesFile: "multiple-repo-keep-latest.yaml",
			input:     manifests,
			// should keep the latest 4 images from both bb8 and redpop
			keepImages: []string{
				"tumblr/bb8:abc123f",             // modified: 0 days ago
				"tumblr/bb8:v0.6.1-261-gbb41394", // modified: 0 days ago
				"tumblr/bb8:v0.6.1-262",          // modified: 0 days ago
				"tumblr/bb8:v0.69.1-262",         // modified: 0 days ago
				"tumblr/redpop:pr-420",           // modified: 1 days ago
				"tumblr/redpop:v1.2.3",           // modified: 4 days ago
				"tumblr/redpop:pr-69",            // modified: 5 days ago
				"tumblr/redpop:pr-69420+13d",     // modified: 13 days ago
			},
			deleteImages: []string{
				"tumblr/bb8:branch-v1.2.3-69",
				"tumblr/bb8:garbage",
				"tumblr/bb8:oldtag-1",
				"tumblr/bb8:oldtag-2",
				"tumblr/bb8:some-ignored-tag",
				"tumblr/bb8:v0.4.2-259-something",
				"tumblr/bb8:v0.5.0-260",
				"tumblr/bb8:v0.5.1-260",
				"tumblr/bb8:v0.5.2",
				"tumblr/bb8:v0.5.23+test",
				"tumblr/bb8:v0.5.3-nice",
				"tumblr/bb8:v0.5.5-420",
				"tumblr/bb8:v0.69-6969",
				"tumblr/redpop:master-2019",
				"tumblr/redpop:master-v1.2.3-69",
				"tumblr/redpop:pr-69419+16d",
				"tumblr/redpop:pr-69420+14d",
				"tumblr/redpop:pr-69421+15d",
				"tumblr/redpop:v1.0.3+metadata",
			},
		},
		{
			rulesFile:    "onlylatest.yaml",
			input:        manifests,
			keepImages:   []string{},
			deleteImages: []string{}, // we dont want to see it in delete, cause its ignored
		},
		{
			rulesFile: "redpop-pr.yaml",
			input:     manifests,
			keepImages: []string{
				"tumblr/redpop:pr-420",
				"tumblr/redpop:pr-69",
				"tumblr/redpop:pr-69420+13d",
				"tumblr/redpop:pr-69420+14d"},
			deleteImages: []string{
				"tumblr/redpop:pr-69419+16d",
				"tumblr/redpop:pr-69421+15d"},
		},
		{
			// ignore some tags that would otherwise get cleaned up by date predicates
			rulesFile:    "bb8-ignore-some.yaml",
			input:        manifests,
			keepImages:   []string{"tumblr/bb8:abc123f"},
			deleteImages: []string{"tumblr/bb8:branch-v1.2.3-69", "tumblr/bb8:garbage", "tumblr/bb8:oldtag-1", "tumblr/bb8:oldtag-2"},
		},
		{
			rulesFile: "bb8-match-version.yaml",
			// this rule should retain only 5 latest version tags
			// and will implicitly skip all versions taht dont parse correctly as a Version
			// meaning there should be no deletedTags that arent correct semantic versions
			input: manifests,
			keepImages: []string{
				"tumblr/bb8:v0.5.23+test",
				"tumblr/bb8:v0.6.1-261-gbb41394",
				"tumblr/bb8:v0.6.1-262",
				"tumblr/bb8:v0.69-6969",
				"tumblr/bb8:v0.69.1-262",
			},
			deleteImages: []string{
				"tumblr/bb8:v0.4.2-259-something",
				"tumblr/bb8:v0.5.0-260",
				"tumblr/bb8:v0.5.1-260",
				"tumblr/bb8:v0.5.2",
				"tumblr/bb8:v0.5.3-nice",
				"tumblr/bb8:v0.5.5-420",
			},
		},
		{
			// this should keep 2 latest version tags, and the last 2 days of all tags.
			// this means there are some versions that would have been deleted, that are still retained
			rulesFile: "bb8-multiple.yaml",
			input:     manifests,
			keepImages: []string{
				"tumblr/bb8:abc123f",
				"tumblr/bb8:v0.69-6969",
				"tumblr/bb8:v0.69.1-262",
			},
			deleteImages: []string{
				"tumblr/bb8:branch-v1.2.3-69",
				"tumblr/bb8:garbage",
				"tumblr/bb8:oldtag-1",
				"tumblr/bb8:oldtag-2",
				"tumblr/bb8:some-ignored-tag", // ignored by 1 rule, but deleted by the nDays rule!
				"tumblr/bb8:v0.4.2-259-something",
				"tumblr/bb8:v0.5.0-260",
				"tumblr/bb8:v0.5.1-260",
				"tumblr/bb8:v0.5.2",
				"tumblr/bb8:v0.5.23+test",
				"tumblr/bb8:v0.5.3-nice",
				"tumblr/bb8:v0.5.5-420",
				"tumblr/bb8:v0.6.1-261-gbb41394",
				"tumblr/bb8:v0.6.1-262",
			},
		},
		{
			rulesFile: "multiple-repo-versions.yaml",
			input:     manifests,
			keepImages: []string{
				"image/x:v0.1.1+x",
				"image/x:v0.6.9+x",
				"image/x:v4.2.1+x",
				"image/y:v0.1.0+y",
				"image/y:v0.69.420+y",
				"image/y:v4.2.0+y",
			},
			deleteImages: []string{
				"image/x:0.0.1+x",
				"image/x:0.0.2+x",
				"image/y:0.0.1+y",
				"image/y:0.0.2+y",
			},
		},
	}
)

type testpayload struct {
	rulesFile    string
	input        []*Manifest
	keepImages   []string
	deleteImages []string
}

func TestApplyRules(t *testing.T) {
	for _, testObj := range tests {
		rulesfile := testObj.rulesFile
		// map intent to a list of versions
		cfg, err := config.LoadFromFile(rulesDir + "/" + rulesfile)
		if err != nil {
			t.Error(err)
			t.Fail()
		}

		keep, delete := ApplyRules(cfg.Rules, testObj.input)
		keep_tags := manifestsAsImageList(keep)
		delete_tags := manifestsAsImageList(delete)
		expectedKeep := testObj.keepImages
		expectedDelete := testObj.deleteImages
		sort.Strings(expectedKeep)
		sort.Strings(expectedDelete)

		if !reflect.DeepEqual(expectedKeep, keep_tags) {
			t.Errorf("%s: expected keep images to be %v but was actually %v", rulesfile, expectedKeep, keep_tags)
			t.Fail()
		}
		if !reflect.DeepEqual(expectedDelete, delete_tags) {
			t.Errorf("%s: expected delete images tags to be %v but was actually %v", rulesfile, expectedDelete, delete_tags)
			t.Fail()
		}
	}
}

func manifestsAsImageList(ms []*Manifest) []string {
	ts := []string{}
	for _, m := range ms {
		ts = append(ts, fmt.Sprintf("%s:%s", m.Name, m.Tag))
	}
	sort.Strings(ts)
	return ts
}
