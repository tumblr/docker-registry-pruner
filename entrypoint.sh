#!/bin/sh
set -e
args=""
if [ -n "$CONFIG" ] ; then
  c="./config/$CONFIG"
  echo "loading configuration from $c"
  args="--config=$c"
fi
set -x
exec docker-registry-pruner $args "$@"
