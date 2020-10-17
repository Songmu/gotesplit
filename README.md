gotesplit
=======

[![Test Status](https://github.com/Songmu/gotesplit/workflows/test/badge.svg?branch=master)][actions]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/gotesplit)][PkgGoDev]

[actions]: https://github.com/Songmu/gotesplit/actions?workflow=test
[license]: https://github.com/Songmu/gotesplit/blob/master/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/gotesplit

gotesplit splits the testng in Go into a subset and run it

## Synopsis

```console
% gotesplit -total=10 -index=0 -- -v -short
go test -v -short -run ^(?:TestAA|TestBB)$
```

## Description

The gotesplit splits the testng in Go into a subset and run it.

It is very useful when you want to run tests in parallel in a CI environment.

## Installation

```console
# Install the latest version. (Install it into ./bin/ by default).
% curl -sfL https://raw.githubusercontent.com/Songmu/gotesplit/master/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
% curl -sfL https://raw.githubusercontent.com/Songmu/gotesplit/master/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
% wget -O - -q https://raw.githubusercontent.com/Songmu/gotesplit/master/install.sh | sh -s [vX.Y.Z]

# go get
% go get github.com/Songmu/gotesplit/cmd/gotesplit
```

## Example

### CircleCI

We don't need to specify the -total and -index flag on CircleCI because gotesplit reads the `CIRCLE_NODE_TOTAL` and `CIRCLE_NODE_INDEX` environment variables automatically.

```yaml
    parallelism: 5
    docker:
      - image: circleci/golang:1.15.3
    steps:
      - checkout
      - run:
          command: |
            curl -sfL https://raw.githubusercontent.com/Songmu/gotesplit/master/install.sh | sh -s
            bin/gotesplit ./... -- -v
```

### GitHub Actions

```
name: CI
on: [push, pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        parallelism: [3]
        index: [0,1,2]
    steps:
      - uses: actions/setup-go@v2
      - uses: actions/checkout@v2
      - name: Run tests parallelly
        run: |
          curl -sfL https://raw.githubusercontent.com/Songmu/gotesplit/master/install.sh | sh -s
          bin/gotesplit -total ${{ matrix.parallelism }} -index ${{ matrix.index }} ./... -- -v
```

## Author

[Songmu](https://github.com/Songmu)
