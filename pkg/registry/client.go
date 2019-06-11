package registry

import (
	"fmt"
	"sync"

	r "github.com/nokia/docker-registry-client/registry"
	"github.com/tumblr/docker-registry-pruner/pkg/config"
	"github.com/tumblr/docker-registry-pruner/pkg/rules"
	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewProduction()
	log       = logger.Sugar()
)

type Client struct {
	r.Registry
	Rules    []*rules.Rule
	nWorkers int
}

type repoTagList struct {
	Repo string
	Tags []string
}

type repoTag struct {
	Repo string
	Tag  string
}

func LogCallback(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func New(c *config.Config) (*Client, error) {
	opts := r.Options{
		Username:      c.Username,
		Password:      c.Password,
		Insecure:      false,
		Logf:          LogCallback,
		DoInitialPing: true,
	}

	hub, err := r.NewCustom(c.RegistryURL, opts)
	if err != nil {
		return nil, err
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	client := Client{
		Registry: *hub,
		Rules:    c.Rules,
		nWorkers: c.Parallelism,
	}
	return &client, nil
}

// given a channel of repos, go get the tags for each one
func (hub *Client) tagFetchWorker(id int, workCh <-chan string, resultCh chan<- repoTagList) {

	for repo := range workCh {
		log.Debugf("%d: looking up tags for %s...", id, repo)
		tags, err := hub.Tags(repo)
		if err != nil {
			log.Warnw("error fetching tags", "repo", repo, "error", err)
			continue
		}
		resultCh <- repoTagList{
			Repo: repo,
			Tags: tags,
		}
	}
	log.Debugf("%d: tag fetcher exiting", id)
}

func (hub *Client) manifestFetchWorker(id int, workCh <-chan repoTag, resultCh chan<- *Manifest) {
	for rt := range workCh {
		log.Debugf("looking up manifest for %s:%s", rt.Repo, rt.Tag)
		m, err := hub.Manifest(rt.Repo, rt.Tag)
		if err != nil {
			// TODO: should we do something terrible here?
			log.Warnw("error fetching manifest", "repo", rt.Repo, "tag", rt.Tag, "error", err)
			continue
		}
		resultCh <- m
	}
	log.Debugf("%d: manifest fetcher exiting", id)
}

func (hub *Client) deleteManifestWorker(id int, workCh <-chan *Manifest, resultCh chan<- error) {
	for m := range workCh {
		log.Infof("%d: deleting manifest for %s:%s", id, m.Name, m.Tag)
		err := hub.DeleteManifest(m)
		if err != nil {
			log.Errorf("%d: error deleting manifest for %s:%s: %v", id, m.Name, m.Tag, err)
			resultCh <- fmt.Errorf("error deleting manifest %s:%s: %v", m.Name, m.Tag, err)
		} else {
			log.Infof("%d: manifest %s:%s successfully deleted", id, m.Name, m.Tag)
			resultCh <- nil
		}
	}
	log.Debugf("%d: delete manifest worker exiting", id)
}

func (hub *Client) RepoTags(repos []string) (map[string][]string, error) {
	var err error
	repositories := repos
	if len(repos) == 0 {
		repositories, err = hub.Repositories()
		if err != nil {
			return nil, err
		}
	}

	// because this is a slow process, lets speed itup by making this a workqueue
	// and use lots of goroutines
	wg := sync.WaitGroup{}
	workCh := make(chan string)
	resultCh := make(chan repoTagList)
	nWorkers := hub.nWorkers
	if nWorkers <= 0 {
		nWorkers = 1
	}
	go func(wg *sync.WaitGroup, resultCh chan repoTagList, workCh chan string) {
		for i := 0; i < hub.nWorkers; i++ {
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				hub.tagFetchWorker(i, workCh, resultCh)
				wg.Done()
			}(i, wg)
		}
		wg.Wait()
		close(resultCh) // signal to consumers there is no more results coming in
	}(&wg, resultCh, workCh)

	go func(workCh chan<- string, repositories []string) {
		// enqueue the work to be done
		for _, repo := range repositories {
			workCh <- repo
		}
		close(workCh) // signal workers
	}(workCh, repositories)

	// read from the results channel and stuff results into our tracking map
	repoTags := map[string][]string{}
	for res := range resultCh {
		repoTags[res.Repo] = res.Tags
	}

	return repoTags, nil
}

func (hub *Client) Manifest(repo, tag string) (*Manifest, error) {
	m, err := hub.ManifestV1(repo, tag)
	if err != nil {
		return nil, err
	}

	return FromSignedManifest(m)
}

func (hub *Client) Manifests(repoTags map[string][]string) ([]*Manifest, error) {
	wg := sync.WaitGroup{}
	workCh := make(chan repoTag)
	resultCh := make(chan *Manifest)
	nWorkers := hub.nWorkers
	if nWorkers <= 0 {
		nWorkers = 1
	}
	go func(wg *sync.WaitGroup, resultCh chan *Manifest, workCh chan repoTag) {
		for i := 0; i < hub.nWorkers; i++ {
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				hub.manifestFetchWorker(i, workCh, resultCh)
				wg.Done()
			}(i, wg)
		}
		wg.Wait()
		close(resultCh) // signal to consumers there is no more results coming in
	}(&wg, resultCh, workCh)

	// enqueue the work to be done
	go func(workCh chan<- repoTag, repoTags map[string][]string) {
		for repo, tags := range repoTags {
			for _, tag := range tags {
				log.Debugf("enqueuing manifest lookup for %s:%s\n", repo, tag)
				workCh <- repoTag{
					Repo: repo,
					Tag:  tag,
				}
			}
		}
		close(workCh) // signal workers
	}(workCh, repoTags)

	// read from the results channel and stuff results into our tracking map
	manifests := []*Manifest{}
	for res := range resultCh {
		manifests = append(manifests, res)
	}

	return manifests, nil
}

func (hub *Client) DeleteManifestsParallel(manifests []*Manifest) (int, []error) {
	// TODO(gabe) we should figure out how to abstract this parallel worker pattern into a generic system

	wg := sync.WaitGroup{}
	workCh := make(chan *Manifest)
	resultCh := make(chan error)
	nWorkers := hub.nWorkers
	if nWorkers <= 0 {
		nWorkers = 1
	}
	go func(wg *sync.WaitGroup, resultCh chan error, workCh chan *Manifest) {
		for i := 0; i < hub.nWorkers; i++ {
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				hub.deleteManifestWorker(i, workCh, resultCh)
				wg.Done()
			}(i, wg)
		}
		wg.Wait()
		close(resultCh) // signal to consumers there is no more results coming in
	}(&wg, resultCh, workCh)

	// enqueue the work to be done
	go func(workCh chan<- *Manifest, manifests []*Manifest) {
		for _, m := range manifests {
			log.Debugf("enqueuing deletion of manifest for %s:%s\n", m.Name, m.Tag)
			workCh <- m
		}
		close(workCh) // signal workers
	}(workCh, manifests)

	errs := []error{}
	deleted := 0
	for err := range resultCh {
		if err != nil {
			errs = append(errs, err)
		} else {
			deleted = deleted + 1
		}
	}

	return deleted, errs
}

func (hub *Client) DeleteManifests(manifests []*Manifest) []error {
	errs := []error{}
	for _, m := range manifests {
		err := hub.DeleteManifest(m)
		if err != nil {
			log.Errorf("unable to delete %s:%s: %v", m.Name, m.Tag, err)
			errs = append(errs, err)
		}
	}
	return errs
}

func (hub *Client) DeleteManifest(m *Manifest) error {
	desc, err := hub.Registry.ManifestDescriptor(m.Name, m.Tag)
	if err != nil {
		return err
	}
	return hub.Registry.DeleteManifest(m.Name, desc.Digest)
}
