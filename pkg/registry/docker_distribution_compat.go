package registry

// see https://docs.docker.com/registry/spec/manifest-v2-1/

// because we want to figure out the relative "age" of an image, the best way is to:
// 1. use the V1 Manifest schema, because it provides some unstructured History
// 2. Snag the latest Created field

import (
	"encoding/json"
	"time"

	"github.com/docker/distribution/manifest/schema1"
)

// internal struct used to extract lastmodified time from a V1 schema
type v1CompatibilityHistory struct {
	Parent  string    `json:"parent,omitempty"`
	Comment string    `json:"comment,omitempty"`
	Created time.Time `json:"created"`
	Config  struct {
		Labels map[string]string `json:"Labels,omitempty"`
	} `json:"config,omitempty"`
	ContainerConfig struct {
		Cmd []string
	} `json:"container_config,omitempty"`
	Author string `json:"author,omitempty"`
}

func deserializeV1CompatibilityHistory(historySerialized string) (v1c v1CompatibilityHistory) {
	// if we cant parse a v1compat struct, just skip suppress the error
	json.Unmarshal([]byte(historySerialized), &v1c)
	return
}

// FromSignedManifest does some parsing and field extraction from the underlying SignedManifest
// and returns our sugar object
func FromSignedManifest(sm *schema1.SignedManifest) (*Manifest, error) {
	// we do some shady stuff to a v1 manifest to extract the latest time any history
	// element was modified. This is used to tell us about an image, and do date-based expiry of images.
	// NOTE: it appears this is not super supported by the docker/distribution API, and is not present
	// in V2 schema! :shruggie:
	// will need to figure out what we want to do for V2 manifests if we want to support date-based expirations

	labels := map[string]string{}
	var lastModified time.Time

	for _, h := range sm.History {
		// keep deserializing the history blobs and extracting any interesting tidbit we can salvage from them
		v1c := deserializeV1CompatibilityHistory(h.V1Compatibility)
		// we care about the most recent Created field
		if v1c.Created.After(lastModified) {
			lastModified = v1c.Created
		}
		// merge all labels found, only adding those that are not already tracked
		if v1c.Config.Labels != nil {
			for k, v := range v1c.Config.Labels {
				if _, ok := labels[k]; !ok {
					labels[k] = v
				}
			}
		}
	}
	return NewManifest(sm.Name, sm.Tag, lastModified, labels)
}
