# docker-registry-pruner 🐳✂️


`docker-registry-pruner` is a rules-based tool that applies business logic to docker images in a Docker Registry storage system for retention.

![GitHub release](https://img.shields.io/github/release/tumblr/docker-registry-pruner.svg) [![Build Status](https://travis-ci.org/tumblr/docker-registry-pruner.svg?branch=master)](https://travis-ci.org/tumblr/docker-registry-pruner) [![Docker Image](https://img.shields.io/badge/Docker%20Hub-tumblr%2Fdocker--registry--pruner-blue.svg)](https://hub.docker.com/r/tumblr/docker-registry-pruner) ![Docker Automated build](https://img.shields.io/docker/automated/tumblr/docker-registry-pruner.svg) ![Docker Build Status](https://img.shields.io/docker/build/tumblr/docker-registry-pruner.svg) ![MicroBadger Size](https://img.shields.io/microbadger/image-size/tumblr/docker-registry-pruner.svg) ![Docker Pulls](https://img.shields.io/docker/pulls/tumblr/docker-registry-pruner.svg) ![Docker Stars](https://img.shields.io/docker/stars/tumblr/docker-registry-pruner.svg) [![Godoc](https://godoc.org/github.com/tumblr/docker-registry-pruner?status.svg)](http://godoc.org/github.com/tumblr/docker-registry-pruner)

# Documentation

## Quickstart

See [configuration overview](/docs/config.md) for how to write a config file. Then, map it into your container and run the report!

```
$ docker run -ti -v $(pwd)/config:/app/config --rm tumblr/docker-registry-pruner --mode report --config ./config/myconfig.yaml
```

Once you are happy with the report, you can perform pruning! WARNING: this is destructive!

```
$ docker run -ti -v $(pwd)/config:/app/config --rm tumblr/docker-registry-pruner --mode prune --config ./config/myconfig.yaml
```

## Configuration

See the [configuration overview](/docs/config.md) for how to write config files to apply retention rules to images in your Registry.

# Examples

Check out [docs/examples.md](/docs/examples.md) for examples using the CLI tool.

# Hacking

See [docs/hacking.md](/docs/hacking.md) for how to hack and contribute.

# License

[Apache 2.0](/LICENSE.txt)

Copyright 2019, Tumblr, Inc.
