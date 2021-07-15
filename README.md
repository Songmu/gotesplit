go-splittestgen
=======
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

go-splittestgen splits test cases into some subsets and print commands to run one of subsets for parallel testing.
* The parser code is based on [github.com/Songmu/gotesplit](https://github.com/Songmu/gotesplit) written by [Songmu](https://github.com/Songmu)

## Usage

```bash
# print test commands
$ go test ./... -list . | go-splittestgen

# execute tests
$ go test ./... -list . | go-splittestgen | sh
```


### Options

```
  -index uint
        index of parallel testing(default 0)
  -total uint
        process num of parallel testing (default 1)
```

## Installation

```bash
# go install
$ go install github.com/minoritea/go-splittestgen/cmd/go-splittestgen
# or just run
$ go test ./... -list . | go run github.com/minoritea/go-splittestgen/cmd/go-splittestgen
```

## Example
### GitHub Actions

```yaml
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
      # Add github.com/minoritea/go-splittestgen to go.mod
      # and install modules before tests.
      - name: Run tests parallelly
        run: |
          go test ./... -list . | \
          go run github.com/minoritea/go-splittestgen/cmd/go-splittestgen | \
            -total ${{ matrix.parallelism }} -index ${{ matrix.index }} | \
          sed -e 's/$/ -v -count 1/g' | sh
```

## Author
[minoritea](https://github.com/minoritea)

## Original Author
[Songmu](https://github.com/Songmu)
