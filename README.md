# tinyini

`tinyini` is a minimalistic library for parsing INI-like configuration
files. For example this is a valid `tinyini` configuration file:

``` ini
globalkey = globalvalue
[section]
key = first-value
key = second-value
empty=
anotherkey = "  has whitespace   "
```
