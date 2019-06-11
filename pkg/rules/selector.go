package rules

import (
	"regexp"
	"sort"
)

type Selector struct {
	Repos []string
	// IgnoreTags will ignore all manifests with the matching tags (regex)
	IgnoreTags []*regexp.Regexp
	// MatchTags will restrict the rule to only apply to manifests matching the regex tag
	MatchTags []*regexp.Regexp
}

func (r *Selector) Match(repo, tag string) bool {
	anyRepoMatch := false
	for _, r := range r.Repos {
		anyRepoMatch = (r == repo) || anyRepoMatch
	}
	if !anyRepoMatch {
		// always return false when this rule does not apply to any listed repos
		return false
	}
	for _, re := range r.IgnoreTags {
		if re.MatchString(tag) {
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
		matchAnyTag = re.MatchString(tag) || matchAnyTag
	}
	return matchAnyTag
}

func MatchAny(selectors []*Selector, repo, tag string) bool {
	anyMatch := false
	for _, selector := range selectors {
		anyMatch = selector.Match(repo, tag) || anyMatch
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

func FilterRepoTags(repoTags map[string][]string, selectors []*Selector) map[string][]string {
	matchingRepoTags := map[string][]string{}
	for repo, tags := range repoTags {
		matchingTags := []string{}
		for _, tag := range tags {
			if MatchAny(selectors, repo, tag) {
				matchingTags = append(matchingTags, tag)
			}
		}
		if len(matchingTags) > 0 {
			sort.Strings(matchingTags)
			matchingRepoTags[repo] = matchingTags
		}
	}
	return matchingRepoTags
}
