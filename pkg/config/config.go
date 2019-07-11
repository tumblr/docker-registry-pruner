package config

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/tumblr/docker-registry-pruner/pkg/rules"
	"gopkg.in/yaml.v2"
)

var (
	// DefaultParallelism default parallelism
	DefaultParallelism = 10
	// ErrMissingRegistry
	ErrMissingRegistry = fmt.Errorf("missing 'registry' key")
	// ErrNoRulesLoaded
	ErrNoRulesLoaded = fmt.Errorf("no rules loaded - did you forget to specify the 'rules' list?")
)

type Config struct {
	RegistryURL  string `yaml:"registry"`
	Username     string
	Password     string
	UsernameFile string `yaml:"username_file"`
	PasswordFile string `yaml:"password_file"`
	Parallelism  int    `yaml:"parallel_workers"`
	// ConfigRules are the loaded rules from the config - these are parsed into actual []rules.Rule
	ConfigRules []*ConfigRule `yaml:"rules"`
	Rules       []*rules.Rule `yaml:"-"`
}

type ConfigRule struct {
	Repos  []string
	Labels map[string]string
	// IgnoreTags will ignore all manifests with the matching tags (regex)
	IgnoreTags []string `yaml:"ignore_tags"`
	// MatchTags will restrict the rule to only apply to manifests matching the regex tag
	MatchTags []string `yaml:"match_tags"`
	// KeepVersions is how many of the latest images to keep, sorted by version
	KeepVersions int `yaml:"keep_versions"`
	// KeepDays is how many days of the images to keep, sorted by last modified
	KeepDays int `yaml:"keep_days"`
	// KeepMostRecent keeps the latest N images, sorted by last modified
	KeepMostRecent int `yaml:"keep_recent"`
}

func LoadFromFile(file string) (*Config, error) {
	c := Config{}

	d, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(d, &c)
	if err != nil {
		return nil, err
	}

	// Support reading username/password from files if present
	if c.UsernameFile != "" {
		s, err := ioutil.ReadFile(c.UsernameFile)
		if err != nil {
			return nil, err
		}
		c.Username = strings.TrimSpace(string(s))
	}
	if c.PasswordFile != "" {
		s, err := ioutil.ReadFile(c.PasswordFile)
		if err != nil {
			return nil, err
		}
		c.Password = strings.TrimSpace(string(s))
	}

	rs, err := rulesFromConfigRules(c.ConfigRules)
	if err != nil {
		return nil, err
	}
	c.Rules = rs

	if c.Parallelism == 0 {
		c.Parallelism = DefaultParallelism
	}

	return &c, c.Validate()
}

func (c *Config) Validate() error {
	if c.RegistryURL == "" {
		return ErrMissingRegistry
	}
	for _, r := range c.Rules {
		err := r.Validate()
		if err != nil {
			return err
		}
	}
	if len(c.Rules) == 0 {
		return ErrNoRulesLoaded
	}
	return nil
}

func rulesFromConfigRules(crs []*ConfigRule) ([]*rules.Rule, error) {
	rules := make([]*rules.Rule, len(crs))
	for i, cr := range crs {
		if !contains(cr.IgnoreTags, "latest") {
			cr.IgnoreTags = append(cr.IgnoreTags, "^latest$")
		}
		r, err := ruleFromConfigRule(cr)
		if err != nil {
			return nil, err
		}
		rules[i] = r
	}
	return rules, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ruleFromConfigRule(cr *ConfigRule) (*rules.Rule, error) {
	r := rules.Rule{
		Selector: rules.Selector{
			Repos:      cr.Repos,
			Labels:     cr.Labels,
			MatchTags:  []*regexp.Regexp{},
			IgnoreTags: []*regexp.Regexp{},
		},
		KeepDays:       cr.KeepDays,
		KeepVersions:   cr.KeepVersions,
		KeepMostRecent: cr.KeepMostRecent,
	}
	if r.Selector.Labels == nil {
		r.Selector.Labels = map[string]string{}
	}
	for _, re := range cr.MatchTags {
		x, err := regexp.Compile(re)
		if err != nil {
			return nil, err
		}
		r.MatchTags = append(r.MatchTags, x)
	}
	for _, re := range cr.IgnoreTags {
		x, err := regexp.Compile(re)
		if err != nil {
			return nil, err
		}
		r.IgnoreTags = append(r.IgnoreTags, x)
	}
	return &r, nil
}
