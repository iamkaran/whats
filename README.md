# [WIP] whats
A CLI Utility that breaks down a command's flags

# Usage
add a `whats` in the front of any command
```
whats $ whats ls -lah
-l    use a long listing format
-a    do not ignore entries starting with .
-h    with -l and -s, print sizes like 1K 234M 2G etc.
whats $
```

# how it works
Using a complex regex pattern, which matches the common format of flags in man pages and --help outputs:

```
       -I, --ignore=PATTERN
              do not list implied entries matching shell PATTERN
```
Which includes the flag line starting with a indent followed by the first flag starting with one or two dashes and then the additional flags separated with a comma or pipe plus a argument hint (=PATTERN).

# installation
`go install https://github.com/iamkaran/whats@latest`
