# whats
A simple cli utility that explains commands using man pages and help text

# usage
add a `whats` in the front of any command
```
$ whats git commit -m "foo"
commit                       Record changes to the repository...
-m <msg>, --message=<msg>    Use <msg> as the commit message. If multiple -m options are given,...

$ whats ls -lahrt
-l                      use a long listing format...
-a, --all               do not ignore entries starting with .
-h, --human-readable    with -l and -s, print sizes like 1K 234M 2G etc.
-r, --reverse           reverse order while sorting...
-t                      sort by time, newest first; see --time...
```

# how it works
Regex pattern matches the common format of flags in man pages and --help outputs:

```
       -I, --ignore=PATTERN
              do not list implied entries matching shell PATTERN
```
Which includes the flag line starting with a indent followed by the first flag starting with one or two dashes and then the additional flags separated with a comma or pipe plus a argument hint (=PATTERN).
Description of sub commands are found by parsing common forms of help text for example `man cmd-subcmd` example: `man git-commit` and these man pages are validated by checking if the synopsis agrees on it being an actual sub command rather than an actual command by the name `cmd-subcmd`.

--help/help is used as a fallback

> [!NOTE]
> Some may complain using regex results in the program being rigid about finding the description but it is the most reliable way right now. (I know man pages can be parsed directly out of their files but every tool likes to be unique about their way of writing it so its better to target the only common thing in all of them which is the format of writing flags/subcommands and their descriptions itself)

# installation
```bash
go install github.com/iamkaran/whats@latest
```

or

```bash
git clone https://github.com/iamkaran/whats.git
cd whats/
go build
```

# roadmap
- [X] Regex pattern for matching flag lines
- [X] Parse man page of binary from argument
- [X] Combined flags (eg: -la)
- [X] Subcommand && it's flags
- [X] Help output as a fallback
- [ ] De-duplicate flags in arguments
