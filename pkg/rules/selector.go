package rules

import (
	"regexp"
	"sort"

	"github.com/tumblr/docker-registry-pruner/pkg/registry"
)

type Selector struct {
	// Repos are a list of repo literal strings that the selector will match
	Repos []string
	// Labels are a map of docker labels that are required to be present on an image to be matched by this selector
	Labels map[string]string
	// IgnoreTags will ignore all manifests with the matching tags (regex)
	IgnoreTags []*regexp.Regexp
	// MatchTags will restrict the rule to only apply to manifests matching the regex tag
	MatchTags []*regexp.Regexp
}

//func (r *Selector) Match(repo, tag string, labels map[string]string) bool {
func (r *Selector) Match(m *registry.Manifest) bool {

	anyRepoMatch := len(r.Repos) == 0 // if r.Repos is empty, assume we have a Repos predicate match
	for _, r := range r.Repos {
		anyRepoMatch = (r == m.Name) || anyRepoMatch
	}
	allLabelsMatch := true // default is that we "match" labels, because empty set is a match
	if len(r.Labels) > 0 {
		// shortcircuit matching if we have a Labels and the image is missing
		// one of our required label keys or values
		// require _all_ labels to match
		for k, v := range r.Labels {
			foundValue, ok := m.Labels[k]
			allLabelsMatch = ok && (foundValue == v) && allLabelsMatch
		}
	}
	if !anyRepoMatch || !allLabelsMatch {
		// require that a Selector match must match any Repos, and if present, all Labels
		// if either of these predicates are not true, bail!
		return false
	}
	for _, re := range r.IgnoreTags {
		if re.MatchString(m.Tag) {
			// always respect ignored tag patterns
			return false
		}
	}
	if len(r.MatchTags) == 0 {
		// if there are no tags to match, return match
		return true
	}
	matchAnyTag := false
	for _, re := range r.MatchTags {
		matchAnyTag = re.MatchString(m.Tag) || matchAnyTag
	}
	return matchAnyTag
}

func MatchAny(selectors []*Selector, m *registry.Manifest) bool {
	anyMatch := false
	for _, selector := range selectors {
		anyMatch = selector.Match(m) || anyMatch
	}
	return anyMatch
}

func RulesToSelectors(ruleset []*Rule) []*Selector {
	ss := make([]*Selector, len(ruleset))
	for i, r := range ruleset {
		ss[i] = &r.Selector
	}
	return ss
}

// FilterManifests will apply a set of Selectors over a slice of Manifests,
// and return the map mapping from repo name to list of matching Manifests.
func FilterManifests(manifests []*registry.Manifest, selectors []*Selector) map[string][]*registry.Manifest {
	matchingManifests := map[string][]*registry.Manifest{}
	for _, manifest := range manifests {
		if MatchAny(selectors, manifest) {
			if _, ok := matchingManifests[manifest.Name]; !ok {
				matchingManifests[manifest.Name] = []*registry.Manifest{}
			}
			matchingManifests[manifest.Name] = append(matchingManifests[manifest.Name], manifest)
		}
	}
	for _, ms := range matchingManifests {
		//sortable := registry.ManifestModifiedCollection(ms)
		sort.Sort(registry.ManifestModifiedCollection(ms))
		//matchingManifests[repo] = sortable
	}
	return matchingManifests
}
