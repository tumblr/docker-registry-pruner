# Config Overview

The config for this is pretty powerful. There is a list of rules, that the rule engine evaluates against the set of Manifests in the Registry. You identify which images to apply the rule to with the `repos:` field. Please note: the repos field is literal string matching, _not_ regex (for now); this is intentional, so as to make cleanup actions explicit rather than implicit in sloppy regex matches.

A Rule is made up of a Selector, and an Action. See below for more details.

## Selectors

A selector is a predicate that images must satisfy to be considered by the `Action` for deletion.

* `repos` is a list of repositories to apply this rule to. This is literal string matching, _not_ regex. (i.e. `tumblr/plumbus`)
* `match_tags` is a list of regexp. Any matching image will have the rule action evaluated against it (i.e. `^v\d+`)
* `ignore_tags` is a list of regexp. Any matching image will explicitly not be evaluated, even if it would have matched `match_tags`

NOTE: the `^latest$` tag is always implicitly inherited into `ignore_tags`.

## Actions

You must provide one action, either `keep_versions`, `keep_recent`, or `keep_days`. Images that match the selector and fail the action predicate will be marked for deletion.

* `keep_versions` (int): Retain the latest N versions of this image, as defined by semantic version ordering. This requires that your tags are properly formatted with semver.org formatting.
* `keep_days` (int): Retain the only images that have been created in the last N days, ordered by image modified date.
* `keep_recent` (int): Retain the latest N images, ordered by when the image was modified date.

NOTE: if your tag does not parse as a valid semantic version, using `keep_versions` can be VERY crazy and best avoided.

NOTE: Any rules are evaluated against the set of tags for a single repo _independently_ from other repos. If you have a rule like the following:

```
- repos:
  - a
  - b
  keep_recent: 5
```

Both `a` and `b` will have 5 images retained, as the rule is evaluated against each repo's set of tags independently.

## Example

```
---
# what registry we should connect to?
registry: https://registry.company.net
# username: <registry username>
# username_file: ./some/file/to/read/containing/username.txt
# password: <registry password>
# password_file: ./some/file/to/read/containing/password.txt

# control parallelism for how queries and deletes are performed in parallel. defaults to 10
# parallel_workers: 10

# selectors to match images, and apply retention logic to them
rules:

  # clean up old versions, but never any tags ending in "release"
  - repos:
      - tumblr/myimage
      - tumblr/anotherimage
    ignore_tags:
      - -release$
    match_tags:
      - ^v\d+.\d+.\d+
    keep_versions: 5

  - repos:
      - tumblr/fleeble
    match_tags:
      - ^v\d+.\d+.\d+
    keep_versions: 5

  # clean up all images with tags matching ^development- after 2wk in these repos
  - repos:
      - some/image
      - another/image
    match_tags:
      - ^development-
    keep_days: 14

  # delete any images for some/image that were built for PRs after 30d
  - repos:
      - some/image
    match_tags:
      - ^pr-\d+
    keep_days: 30

  # keep only the most recent 5 images by modification time
  - repos:
      - web/devtools
    keep_recent: 5
```

