# lhdiff

A Lightweight Hybrid Approach for Tracking Source Lines

[![Go Report Card](https://goreportcard.com/badge/github.com/aslakhellesoy/lhdiff)](https://goreportcard.com/report/github.com/aslakhellesoy/lhdiff)
[![Coverage Status](https://img.shields.io/codecov/c/github/aslakhellesoy/lhdiff.svg)](https://codecov.io/gh/aslakhellesoy/lhdiff)
[![Release](https://github.com/aslakhellesoy/lhdiff/workflows/Release/badge.svg)](https://github.com/aslakhellesoy/lhdiff/releases)

## Install

    go get github.com/aslakhellesoy/lhdiff

To install from source, see [CONTRIBUTING.md](./CONTRIBUTING.md)

## Usage

    lhdiff left right

Example using git:

    dist/lhdiff_darwin_amd64/lhdiff \
    <( git show 400a62e39d39d231d8160002dfb7ed95a004278b:cmd/lhdiff/main.go ) \
    <( git show 35f1ba7b554d69a07e59d6f69297d08599f4217c:cmd/lhdiff/main.go )

    dist/lhdiff_darwin_amd64/lhdiff \
    <( git show 085519173c4e6e76c425dac0a628f21ff0cdcfa8:lhdiff.go ) \
    <( git show 4ae3495de0c31675940861592a3929df8154785f:lhdiff.go )

# Related

* [Incremental Origin Analysis of Source Code Files](http://citeseerx.ist.psu.edu/viewdoc/download?doi=10.1.1.721.548&rep=rep1&type=pdf)

# LICENSE

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg)](http://www.opensource.org/licenses/MIT)

This is distributed under the [MIT License](http://www.opensource.org/licenses/MIT).
