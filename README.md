# lhdiff

A Lightweight Hybrid Approach for Tracking Source Lines.

Lhdiff provides a mapping of individual lines between two revisions of the same file. It provides a better mapping than
Unix diff, and works independently of the file contents (programming language). See [ARCHITECTURE.md](ARCHITECTURE.md) for more details.

## Install

    go get github.com/SmartBear/lhdiff

To install from source, see [CONTRIBUTING.md](./CONTRIBUTING.md)

## Usage

There are two ways to use lhdiff - as a commandline program, or as a library

### Command line

    lhdiff [--compact] left right

Example using git:

    lhdiff --compact \
    <( git show 400a62e39d39d231d8160002dfb7ed95a004278b:cmd/lhdiff/main.go ) \
    <( git show 35f1ba7b554d69a07e59d6f69297d08599f4217c:cmd/lhdiff/main.go )

    lhdiff --compact \
    <( git show 085519173c4e6e76c425dac0a628f21ff0cdcfa8:lhdiff.go ) \
    <( git show 4ae3495de0c31675940861592a3929df8154785f:lhdiff.go )

### Library

```go
left := `one two three four
eight
nine ten eleven twelve
thirteen fourteen fifteen`

right := `one two three four
nine ten twelve
five six BANANA seven eight
APPLE PEAR
thirteen fourteen fifteen
`

linePairs := Lhdiff(left, right, 4)
PrintLinePairs(linePairs, false)

// Output:
// 1,1
// 2,_
// 3,2
// 4,5
```

# Related

* [diffsitter](https://github.com/afnanenayet/diffsitter)
* [Incremental Origin Analysis of Source Code Files](http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.721.548&rep=rep1&type=pdf)

# LICENSE

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://www.opensource.org/licenses/MIT)

This is distributed under the [MIT License](http://www.opensource.org/licenses/MIT).
