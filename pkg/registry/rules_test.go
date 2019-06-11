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
		mkmanifest("tumblr/plumbus", "v1.2.3", 4),
		mkmanifest("tumblr/plumbus", "v1.0.3+metadata", 69),
		mkmanifest("tumblr/plumbus", "pr-69420+13d", 13),
		mkmanifest("tumblr/plumbus", "pr-69420+14d", 14),
		mkmanifest("tumblr/plumbus", "master-v1.2.3-69", 14),
		mkmanifest("tumblr/plumbus", "master-2019", 14),
		mkmanifest("tumblr/plumbus", "pr-420", 1),
		mkmanifest("tumblr/plumbus", "pr-69", 5),
		mkmanifest("tumblr/plumbus", "pr-69421+15d", 15),
		mkmanifest("tumblr/plumbus", "pr-69419+16d", 16),
		mkmanifest("image/latest", "latest", 69420),
		mkmanifest("tumblr/fleeble", "latest", 0),
		mkmanifest("tumblr/fleeble", "garbage", 5),
		mkmanifest("tumblr/fleeble", "v0.4.2-259-something", 5),
		mkmanifest("tumblr/fleeble", "v0.5.0-260", 4),
		mkmanifest("tumblr/fleeble", "v0.5.1-260", 4),
		mkmanifest("tumblr/fleeble", "v0.5.23+test", 3),
		mkmanifest("tumblr/fleeble", "v0.5.2", 3),
		mkmanifest("tumblr/fleeble", "some-ignored-tag", 69),
		mkmanifest("tumblr/fleeble", "oldtag-1", 69),
		mkmanifest("tumblr/fleeble", "oldtag-2", 70),
		mkmanifest("tumblr/fleeble", "v0.6.1-261-gbb41394", 0),
		mkmanifest("tumblr/fleeble", "v0.5.3-nice", 1),
		mkmanifest("tumblr/fleeble", "v0.69-6969", 1),
		mkmanifest("tumblr/fleeble", "v0.5.5-420", 1),
		mkmanifest("tumblr/fleeble", "v0.6.1-262", 0),
		mkmanifest("tumblr/fleeble", "branch-v1.2.3-69", 14),
		mkmanifest("tumblr/fleeble", "v0.69.1-262", 0),
		mkmanifest("tumblr/fleeble", "abc123f", 0),

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
			// should keep the latest 4 images from both fleeble and plumbus
			keepImages: []string{
				"tumblr/fleeble:abc123f",             // modified: 0 days ago
				"tumblr/fleeble:v0.6.1-261-gbb41394", // modified: 0 days ago
				"tumblr/fleeble:v0.6.1-262",          // modified: 0 days ago
				"tumblr/fleeble:v0.69.1-262",         // modified: 0 days ago
				"tumblr/plumbus:pr-420",              // modified: 1 days ago
				"tumblr/plumbus:v1.2.3",              // modified: 4 days ago
				"tumblr/plumbus:pr-69",               // modified: 5 days ago
				"tumblr/plumbus:pr-69420+13d",        // modified: 13 days ago
			},
			deleteImages: []string{
				"tumblr/fleeble:branch-v1.2.3-69",
				"tumblr/fleeble:garbage",
				"tumblr/fleeble:oldtag-1",
				"tumblr/fleeble:oldtag-2",
				"tumblr/fleeble:some-ignored-tag",
				"tumblr/fleeble:v0.4.2-259-something",
				"tumblr/fleeble:v0.5.0-260",
				"tumblr/fleeble:v0.5.1-260",
				"tumblr/fleeble:v0.5.2",
				"tumblr/fleeble:v0.5.23+test",
				"tumblr/fleeble:v0.5.3-nice",
				"tumblr/fleeble:v0.5.5-420",
				"tumblr/fleeble:v0.69-6969",
				"tumblr/plumbus:master-2019",
				"tumblr/plumbus:master-v1.2.3-69",
				"tumblr/plumbus:pr-69419+16d",
				"tumblr/plumbus:pr-69420+14d",
				"tumblr/plumbus:pr-69421+15d",
				"tumblr/plumbus:v1.0.3+metadata",
			},
		},
		{
			rulesFile:    "onlylatest.yaml",
			input:        manifests,
			keepImages:   []string{},
			deleteImages: []string{}, // we dont want to see it in delete, cause its ignored
		},
		{
			rulesFile: "plumbus-pr.yaml",
			input:     manifests,
			keepImages: []string{
				"tumblr/plumbus:pr-420",
				"tumblr/plumbus:pr-69",
				"tumblr/plumbus:pr-69420+13d",
				"tumblr/plumbus:pr-69420+14d"},
			deleteImages: []string{
				"tumblr/plumbus:pr-69419+16d",
				"tumblr/plumbus:pr-69421+15d"},
		},
		{
			// ignore some tags that would otherwise get cleaned up by date predicates
			rulesFile:    "fleeble-ignore-some.yaml",
			input:        manifests,
			keepImages:   []string{"tumblr/fleeble:abc123f"},
			deleteImages: []string{"tumblr/fleeble:branch-v1.2.3-69", "tumblr/fleeble:garbage", "tumblr/fleeble:oldtag-1", "tumblr/fleeble:oldtag-2"},
		},
		{
			rulesFile: "fleeble-match-version.yaml",
			// this rule should retain only 5 latest version tags
			// and will implicitly skip all versions taht dont parse correctly as a Version
			// meaning there should be no deletedTags that arent correct semantic versions
			input: manifests,
			keepImages: []string{
				"tumblr/fleeble:v0.5.23+test",
				"tumblr/fleeble:v0.6.1-261-gbb41394",
				"tumblr/fleeble:v0.6.1-262",
				"tumblr/fleeble:v0.69-6969",
				"tumblr/fleeble:v0.69.1-262",
			},
			deleteImages: []string{
				"tumblr/fleeble:v0.4.2-259-something",
				"tumblr/fleeble:v0.5.0-260",
				"tumblr/fleeble:v0.5.1-260",
				"tumblr/fleeble:v0.5.2",
				"tumblr/fleeble:v0.5.3-nice",
				"tumblr/fleeble:v0.5.5-420",
			},
		},
		{
			// this should keep 2 latest version tags, and the last 2 days of all tags.
			// this means there are some versions that would have been deleted, that are still retained
			rulesFile: "fleeble-multiple.yaml",
			input:     manifests,
			keepImages: []string{
				"tumblr/fleeble:abc123f",
				"tumblr/fleeble:v0.69-6969",
				"tumblr/fleeble:v0.69.1-262",
			},
			deleteImages: []string{
				"tumblr/fleeble:branch-v1.2.3-69",
				"tumblr/fleeble:garbage",
				"tumblr/fleeble:oldtag-1",
				"tumblr/fleeble:oldtag-2",
				"tumblr/fleeble:some-ignored-tag", // ignored by 1 rule, but deleted by the nDays rule!
				"tumblr/fleeble:v0.4.2-259-something",
				"tumblr/fleeble:v0.5.0-260",
				"tumblr/fleeble:v0.5.1-260",
				"tumblr/fleeble:v0.5.2",
				"tumblr/fleeble:v0.5.23+test",
				"tumblr/fleeble:v0.5.3-nice",
				"tumblr/fleeble:v0.5.5-420",
				"tumblr/fleeble:v0.6.1-261-gbb41394",
				"tumblr/fleeble:v0.6.1-262",
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
