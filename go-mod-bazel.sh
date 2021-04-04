#!/bin/sh
set -e

go mod tidy

if [ -f go.sum ]; then
  # Temporary solution to workaround https://github.com/bazelbuild/bazel-gazelle/issues/990
  git add go.sum
fi

bazel run //:gazelle -- update-repos -from_file=go.mod -prune -to_macro=external.bzl%go_dependencies
bazel run //:gazelle -- fix

if [ -f go.sum ]; then
  # Temporary solution to workaround https://github.com/bazelbuild/bazel-gazelle/issues/990
  git restore go.sum
fi
