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
type v1Compatibility struct {
	Parent          string    `json:"parent,omitempty"`
	Comment         string    `json:"comment,omitempty"`
	Created         time.Time `json:"created"`
	ContainerConfig struct {
		Cmd []string
	} `json:"container_config,omitempty"`
	Author string `json:"author,omitempty"`
}

// lastModified does some shady stuff to a v1 manifest to extract the latest time any history
// element was modified. This is used to tell us about an image, and do date-based expiry of images.
// NOTE: it appears this is not super supported by the docker/distribution API, and is not present
// in V2 schema! :shruggie:
// will need to figure out what we want to do for V2 manifests if we want to support date-based expirations
func lastModified(m *schema1.SignedManifest) time.Time {
	var t time.Time
	for _, h := range m.History {
		v1c := v1Compatibility{}
		err := json.Unmarshal([]byte(h.V1Compatibility), &v1c)
		if err != nil {
			// if we cant parse a v1compat struct, just skip
			continue
		}
		if v1c.Created.After(t) {
			t = v1c.Created
		}
	}
	return t
}
