package registry

// ManifestVersionCollection is a type that implements the sort.Interface interface
// so that versions can be sorted.
type ManifestVersionCollection []*Manifest

func (v ManifestVersionCollection) Len() int {
	return len(v)
}

func (v ManifestVersionCollection) Less(i, j int) bool {
	return v[i].Version.LessThan(v[j].Version)
}

func (v ManifestVersionCollection) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// ManifestModifiedCollection is a type that implements the sort.Interface interface
// so that Manifests can be sorted by modification time
type ManifestModifiedCollection []*Manifest

func (v ManifestModifiedCollection) Len() int {
	return len(v)
}

func (v ManifestModifiedCollection) Less(i, j int) bool {
	return v[i].LastModified.Before(v[j].LastModified)
}

func (v ManifestModifiedCollection) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
