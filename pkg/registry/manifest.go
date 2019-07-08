package registry

import (
	"fmt"
	"regexp"
	//"sort"
	"time"

	"github.com/hashicorp/go-version"
	"go.uber.org/zap"
	//"github.com/tumblr/docker-registry-pruner/pkg/rules"
)

var (
	// DefaultVersion is the default version we use for a Manifest if we cant parse it
	DefaultVersion = version.Must(version.NewVersion("0.0.0"))
	// GitShaRegex is the anchored regex that a pure commit sha matches
	GitShaRegex = regexp.MustCompile(`^[0-9a-f]{4,}$`)
	logger, _   = zap.NewProduction()
	log         = logger.Sugar()
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
	Labels  map[string]string
}

// NewManifest creates a new Manifest
func NewManifest(repo string, tag string, lm time.Time, labels map[string]string) (*Manifest, error) {
	var err error
	mani := Manifest{
		Name:         repo,
		Tag:          tag,
		LastModified: lm,
		Labels:       labels,
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

// removes all items in b from a, returning the list (a-b)
// this is super shitty timecomplexity but i really dont care
func RemoveItems(a []*Manifest, b []*Manifest) []*Manifest {
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

// DedupeManifests will deduplicate a list of Manifests by name:tag
func DedupeManifests(s []*Manifest) []*Manifest {
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
