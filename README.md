# tinyini

`tinyini` is a minimalistic library for parsing INI-like configuration
files.

[![Tests](https://img.shields.io/github/workflow/status/susji/tinyini/Go?label=tests)](https://github.com/susji/tinyini/actions/workflows/go.yml)
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg?label=pkg.go.dev)](https://pkg.go.dev/github.com/susji/tinyini)

## example configuration file

``` {.ini}
globalkey = globalvalue

[section]
key = first-value
key = second-value
empty=
anotherkey = "  has whitespace   "

[änöther-section] ; this is a comment and ignored
key = different value
```
