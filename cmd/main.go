package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/tumblr/docker-registry-pruner/pkg/config"
	"github.com/tumblr/docker-registry-pruner/pkg/registry"
	"github.com/tumblr/docker-registry-pruner/pkg/rules"
	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewProduction()
	log       = logger.Sugar()
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		ShowVersion(os.Stderr)
		flag.PrintDefaults()
	}
}

func main() {
	defer logger.Sync()

	var (
		configFile string
		mode       string
	)
	flag.StringVar(&configFile, "config", "config.yaml", "Config yaml")
	flag.StringVar(&mode, "mode", "report", "Select operation mode")
	flag.Parse()

	cfg, err := config.LoadFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	hub, err := registry.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Created Registry client for %s", cfg.RegistryURL)

	// make a list of unique repos we are gonna lookup from the config's rules
	reposMap := map[string]bool{}
	for _, cr := range cfg.Rules {
		for _, r := range cr.Repos {
			reposMap[r] = true
		}
	}
	repos := []string{}
	for repo := range reposMap {
		repos = append(repos, repo)
	}

	for _, rule := range hub.Rules {
		log.Infof("Loaded rule: %s", rule.String())
	}

	switch mode {
	case "report":
		log.Infof("Building image report for images: %s", strings.Join(repos, ", "))
		ShowMatchingRepos(hub, repos)
	case "prune":
		log.Infof("Pruning tags for images: %s", strings.Join(repos, ", "))
		ok := DeleteMatchingImages(hub, repos)
		if !ok {
			os.Exit(2)
		}
	default:
		log.Fatalf("Unsupported mode %s", mode)
	}

}

func PrintTableManifests(matches map[string][]*registry.Manifest) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "action\timage\ttag\tparsed_version\tage_days\n")
	for action, manifests := range matches {
		for _, m := range manifests {
			daysOld := int64(time.Since(m.LastModified).Hours() / 24.0)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", action, m.Name, m.Tag, m.Version.String(), daysOld)
		}
	}
	w.Flush()
}

func FetchImagesAndApplyRules(hub *registry.Client, repos []string) map[string][]*registry.Manifest {
	repoTags, err := hub.RepoTags(repos)
	if err != nil {
		log.Fatal(err)
	}

	selectors := rules.RulesToSelectors(hub.Rules)
	filteredRepoTags := rules.FilterRepoTags(repoTags, selectors)
	for repo, tags := range filteredRepoTags {
		log.Debugf("Repo %s has %d tag matching ruless\n", repo, len(tags))
	}

	allManifests, err := hub.Manifests(filteredRepoTags)
	if err != nil {
		log.Fatal(err)
	}

	keep, delete := registry.ApplyRules(hub.Rules, allManifests)
	matches := map[string][]*registry.Manifest{
		"keep":   keep,
		"delete": delete,
	}
	return matches
}

func ShowMatchingRepos(hub *registry.Client, repos []string) {
	log.Infof("Querying for manifests. This may take a while...")
	matches := FetchImagesAndApplyRules(hub, repos)
	deletes, keeps := len(matches["delete"]), len(matches["keep"])
	PrintTableManifests(matches)
	fmt.Fprintf(os.Stderr, "deleting %d images, keeping %d images\n", deletes, keeps)
}

func DeleteMatchingImages(hub *registry.Client, repos []string) bool {
	log.Infof("Querying for manifests. This may take a while...")
	matches := FetchImagesAndApplyRules(hub, repos)
	log.Infof("Beginning deletion of %d images", len(matches["delete"]))
	deleted, errs := hub.DeleteManifestsParallel(matches["delete"])
	log.Infof("Deleted %d images, encountered %d errors", deleted, len(errs))
	return len(errs) == 0
}
