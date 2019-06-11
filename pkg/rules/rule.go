package rules

import (
	"fmt"
	"strings"
)

var (
	// ErrKeepVersionsMustBePositive
	ErrKeepVersionsMustBePositive = fmt.Errorf("keep_versions must be positive")
	// ErrKeepDaysMustBePositive
	ErrKeepDaysMustBePositive = fmt.Errorf("keep_days must be positive")
	// ErrKeepMostRecentCountMustBePositive
	ErrKeepMostRecentCountMustBePositive = fmt.Errorf("keep_recent must be positive")
	// ErrMissingRepos
	ErrMissingRepos = fmt.Errorf("repos field missing")
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
	return fmt.Sprintf("Repos:%s Selector{%s} Action{%s}", strings.Join(r.Repos, ","), selector, action)
}

func (r *Rule) Validate() error {
	switch {
	case len(r.Repos) == 0:
		return ErrMissingRepos
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
