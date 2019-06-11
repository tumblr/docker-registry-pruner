package registry

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/hashicorp/go-version"
	"github.com/tumblr/docker-registry-pruner/pkg/rules"
)

var (
	// DefaultVersion is the default version we use for a Manifest if we cant parse it
	DefaultVersion = version.Must(version.NewVersion("0.0.0"))
	// GitShaRegex is the anchored regex that a pure commit sha matches
	GitShaRegex = regexp.MustCompile(`^[0-9a-f]{4,}$`)
)

// Manifest is a combined struct of a v1 manifest, as well as some interesting fields
// we layer on top.
type Manifest struct {
	Name string
	Tag  string
	//FSLayers []*schema1.FSLayer
	// LastModified is a synthesized field we extract from History via `lastModified`
	LastModified time.Time
	// Version is a sortable version field, derived from Tag
	Version *version.Version
}

func must(m *Manifest, err error) *Manifest {
	if err != nil {
		panic(err)
	}
	return m
}

// FromSignedManifest does some parsing and field extraction from the underlying SignedManifest
// and returns our sugar object
func FromSignedManifest(sm *schema1.SignedManifest) (*Manifest, error) {
	return NewManifestWithLastModified(sm.Name, sm.Tag, lastModified(sm))
}

// NewManifest creates a new Manifest
func NewManifest(repo string, tag string) (*Manifest, error) {
	return NewManifestWithLastModified(repo, tag, time.Unix(0, 0))
}

// NewManifestWithLastModified creates a new Manifest
func NewManifestWithLastModified(repo string, tag string, lm time.Time) (*Manifest, error) {
	var err error
	mani := Manifest{
		Name:         repo,
		Tag:          tag,
		LastModified: lm,
	}

	// do some version parsing of the tag, as well!
	// before we do any version parsing, lets match against a raw git sha - hashicorp version parsing produces nonsense
	// values when parsing shas, so lets skip this and just assume DefaultVersion here.
	if GitShaRegex.MatchString(tag) {
		log.Debugf("Assuming default version, because %s is just a git sha", tag)
		mani.Version = DefaultVersion
	} else {

		mani.Version, err = version.NewVersion(tag)
		if err != nil {
			// lets make the assumption that we just take the minimal version
			mani.Version = DefaultVersion
		}
	}

	return &mani, nil
}

// Match tells whether a Selector matches thsi manifest. It uses the ignore* and match*
// fields of the selector.
func (m *Manifest) Match(selector rules.Selector) bool {
	return selector.Match(m.Name, m.Tag)
}

// ApplyRules takes a list of rules, and applies them to a list of manifests.
// 2 stages: 1. matching selectors, 2. of those that match, apply retention logic in rule
// returns 2 slices; the manifests to keep, and those to delete
func ApplyRules(ruleset []*rules.Rule, manifests []*Manifest) (keep []*Manifest, delete []*Manifest) {
	manifestsByRepo := map[string][]*Manifest{}
	// group manifests by their repo, so we apply rule sets only over one repo's manifests at a time
	for _, manifest := range manifests {
		ms, ok := manifestsByRepo[manifest.Name]
		if !ok {
			ms = []*Manifest{}
		}
		manifestsByRepo[manifest.Name] = append(ms, manifest)
	}

	// apply rules to manifests
	for _, manifests := range manifestsByRepo {
		k, d := applyRules(ruleset, manifests)
		/*
			for _, r := range ruleset {
				fmt.Printf("rules: %+v\n", *r)
			}
			fmt.Printf("got keep=%v\n", manifestsAsTagList(k))
			fmt.Printf("got delete=%v\n", manifestsAsTagList(d))
		*/
		keep = append(keep, k...)
		delete = append(delete, d...)
	}

	// 3. dedupe our keep/delete sets, because we definitely could have matched an image with multiple rules
	// NOTE: delete supercedes any keep directive, because keep is a default.
	// TODO: we will need to remove all the deletes from keeps
	keep = removeItems(keep, delete)
	return dedupeManifests(keep), dedupeManifests(delete)
}

// applyRules returns a list of Manifests that match the set of rules
// assumes all manifests are for the same repo!
func applyRules(ruleset []*rules.Rule, manifests []*Manifest) (keep []*Manifest, delete []*Manifest) {
	for _, rule := range ruleset {
		// 1. for each rule, see if any manifests match our selector.
		filteredManifests := []*Manifest{}
		for _, manifest := range manifests {
			// see if this rule's Selector matches any of these images for _, manifest := range manifests {
			if manifest.Match(rule.Selector) {
				filteredManifests = append(filteredManifests, manifest)
			}
		}

		// 2. For all manifests that were selected by this rule, apply retention logic to it
		switch {
		case rule.KeepVersions > 0:
			// handle versions that arent parsable. We do not apply any retention rules to versions that didnt parse
			validVersionManifests := []*Manifest{}
			for _, manifest := range filteredManifests {
				if manifest.Version != DefaultVersion {
					validVersionManifests = append(validVersionManifests, manifest)
				}
			}
			sort.Sort(ManifestVersionCollection(validVersionManifests))
			indexHigh := len(validVersionManifests)
			indexLow := indexHigh - rule.KeepVersions
			if indexLow < 0 {
				indexLow = 0
			}
			delete = append(delete, validVersionManifests[0:indexLow]...)
			keep = append(keep, validVersionManifests[indexLow:indexHigh]...)

		case rule.KeepDays > 0:
			tNow := time.Now()
			sort.Sort(ManifestModifiedCollection(filteredManifests))
			for _, manifest := range filteredManifests {
				if int64(tNow.Sub(manifest.LastModified).Minutes()) > int64(24*60*rule.KeepDays) {
					delete = append(delete, manifest)
				} else {
					keep = append(keep, manifest)
				}
			}
		case rule.KeepMostRecent > 0:
			sort.Sort(ManifestModifiedCollection(filteredManifests))
			for i, manifest := range filteredManifests {
				if i < len(filteredManifests)-rule.KeepMostRecent {
					delete = append(delete, manifest)
				} else {
					keep = append(keep, manifest)
				}
			}
		}
	}
	return
}

// removes all items in b from a, returning the list (a-b)
// this is super shitty timecomplexity but i really dont care
func removeItems(a []*Manifest, b []*Manifest) []*Manifest {
	newa := make([]*Manifest, len(a))
	for i, x := range a {
		newa[i] = x
	}
	for _, del := range b {
		for i := 0; i < len(newa); i++ {
			if newa[i].Name == del.Name && newa[i].Tag == del.Tag {
				newa = append(newa[:i], newa[i+1:]...)
				i-- // make sure we properly splice out the element repecting index
			}
		}
	}
	return newa
}

func dedupeManifests(s []*Manifest) []*Manifest {
	seen := make(map[string]struct{}, len(s))
	j := 0
	for _, v := range s {
		k := fmt.Sprintf("%s:%s", v.Name, v.Tag)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}
