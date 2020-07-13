package rules

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tumblr/docker-registry-pruner/pkg/registry"
)

var (
	// ErrLabelsNil is returned when an initialization error creates a Selector with a nil Labels map
	ErrLabelsNil = fmt.Errorf("labels must not be a nil map")
	// ErrKeepVersionsMustBePositive
	ErrKeepVersionsMustBePositive = fmt.Errorf("keep_versions must be positive")
	// ErrKeepDaysMustBePositive
	ErrKeepDaysMustBePositive = fmt.Errorf("keep_days must be positive")
	// ErrKeepMostRecentCountMustBePositive
	ErrKeepMostRecentCountMustBePositive = fmt.Errorf("keep_recent must be positive")
	// ErrMissingReposOrLabels
	ErrMissingReposOrLabels = fmt.Errorf("repos or labels selector is required")
	// ErrActionMustBeSpecified
	ErrActionMustBeSpecified        = fmt.Errorf("one of keep_versions, keep_days, or keep_recent must be specified as an action")
	ErrMultipleActionVersionsDays   = fmt.Errorf("both keep_versions and keep_days specified, but are mutually exclusive")
	ErrMultipleActionDaysLatest     = fmt.Errorf("both keep_days and keep_recent specified, but are mutually exclusive")
	ErrMultipleActionLatestVersions = fmt.Errorf("both keep_versions and keep_recent specified, but are mutually exclusive")
)

type Rule struct {
	Selector

	// KeepVersions is how many of the latest images to keep, sorted by version
	KeepVersions int
	// KeepDays is how many of the latest images to keep, sorted by last modified
	KeepDays int
	// KeepMostRecent will keep the latest N images, by modification time
	KeepMostRecent int
}

// String returns a useful string description of this Rule
func (r *Rule) String() string {
	ignores := []string{}
	for _, i := range r.IgnoreTags {
		ignores = append(ignores, i.String())
	}
	matches := []string{}
	for _, i := range r.MatchTags {
		matches = append(matches, i.String())
	}
	selector := fmt.Sprintf("ignore tags [%s], match tags [%s]", strings.Join(ignores, " or "), strings.Join(matches, " or "))
	action := ""
	if r.KeepMostRecent != 0 {
		action = fmt.Sprintf("keep latest %d images", r.KeepMostRecent)
	}
	if r.KeepDays != 0 {
		action = fmt.Sprintf("keep latest %d days", r.KeepDays)
	}
	if r.KeepVersions != 0 {
		action = fmt.Sprintf("keep latest %d versions", r.KeepVersions)
	}
	return fmt.Sprintf("Repos:%s Labels:%v Selector{%s} Action{%s}", strings.Join(r.Repos, ","), r.Labels, selector, action)
}

func (r *Rule) Validate() error {
	switch {
	case r.Labels == nil:
		return ErrLabelsNil
	case len(r.Repos) == 0 && len(r.Labels) == 0:
		return ErrMissingReposOrLabels
	case r.KeepDays != 0 && r.KeepVersions != 0:
		return ErrMultipleActionVersionsDays
	case r.KeepDays != 0 && r.KeepMostRecent != 0:
		return ErrMultipleActionDaysLatest
	case r.KeepMostRecent != 0 && r.KeepVersions != 0:
		return ErrMultipleActionLatestVersions
	case r.KeepDays < 0:
		return ErrKeepDaysMustBePositive
	case r.KeepVersions < 0:
		return ErrKeepVersionsMustBePositive
	case r.KeepMostRecent < 0:
		return ErrKeepMostRecentCountMustBePositive
	case r.KeepDays == 0 && r.KeepVersions == 0 && r.KeepMostRecent == 0:
		return ErrActionMustBeSpecified
	default:
		return nil
	}
}

// ApplyRules takes a list of rules, and applies them to a list of manifests.
// 2 stages: 1. matching selectors, 2. of those that match, apply retention logic in rule
// returns 2 slices; the manifests to keep, and those to delete
func ApplyRules(ruleset []*Rule, manifests []*registry.Manifest) (keep []*registry.Manifest, delete []*registry.Manifest) {
	manifestsByRepo := map[string][]*registry.Manifest{}
	// group manifests by their repo, so we apply rule sets only over one repo's manifests at a time
	for _, manifest := range manifests {
		ms, ok := manifestsByRepo[manifest.Name]
		if !ok {
			ms = []*registry.Manifest{}
		}
		manifestsByRepo[manifest.Name] = append(ms, manifest)
	}

	// apply rules to manifests
	for _, manifests := range manifestsByRepo {
		k, d := applyRules(ruleset, manifests)
		keep = append(keep, k...)
		delete = append(delete, d...)
	}

	// 3. dedupe our keep/delete sets, because we definitely could have matched an image with multiple rules
	// NOTE: delete supercedes any keep directive, because keep is a default.
	// TODO: we will need to remove all the deletes from keeps
	keep = registry.RemoveItems(keep, delete)
	return registry.DedupeManifests(keep), registry.DedupeManifests(delete)
}

// applyRules returns a list of Manifests that match the set of rules
// assumes all manifests are for the same repo!
func applyRules(ruleset []*Rule, manifests []*registry.Manifest) (keep []*registry.Manifest, delete []*registry.Manifest) {
	for _, rule := range ruleset {
		// 1. for each rule, see if any manifests match our selector.
		filteredManifests := []*registry.Manifest{}
		for _, manifest := range manifests {
			// see if this rule's Selector matches any of these manifests
			if rule.Match(manifest) {
				filteredManifests = append(filteredManifests, manifest)
			}
		}

		// 2. For all manifests that were selected by this rule, apply retention logic to it
		switch {
		case rule.KeepVersions > 0:
			// handle versions that arent parsable. We do not apply any retention rules to versions that didnt parse
			validVersionManifests := []*registry.Manifest{}
			for _, manifest := range filteredManifests {
				if manifest.Version != registry.DefaultVersion {
					validVersionManifests = append(validVersionManifests, manifest)
				}
			}
			sort.Reverse(registry.ManifestVersionCollection(validVersionManifests))
			indexHigh := len(validVersionManifests)
			indexLow := indexHigh - rule.KeepVersions
			if indexLow < 0 {
				indexLow = 0
			}
			delete = append(delete, validVersionManifests[0:indexLow]...)
			keep = append(keep, validVersionManifests[indexLow:indexHigh]...)

		case rule.KeepDays > 0:
			tNow := time.Now()
			sort.Sort(registry.ManifestModifiedCollection(filteredManifests))
			for _, manifest := range filteredManifests {
				if int64(tNow.Sub(manifest.LastModified).Minutes()) > int64(24*60*rule.KeepDays) {
					delete = append(delete, manifest)
				} else {
					keep = append(keep, manifest)
				}
			}
		case rule.KeepMostRecent > 0:
			sort.Sort(registry.ManifestModifiedCollection(filteredManifests))
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
