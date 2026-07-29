# whats
A simple cli utility that explains commands using man pages and help text
<img width="700" height="300" alt="demo" src="https://github.com/user-attachments/assets/6e16631e-443a-4dac-a5aa-1c0b6bb44bcc" />
# usage
add a `whats` in the front of any command
```
$ whats git commit -m "foo"
commit                       Record changes to the repository..:
-m <msg>, --message=<msg>    Use <msg> as the commit message. If multiple -m options are given,...

$ whats ls -lahrt
-l                      use a long listing format...
-a, --all               do not ignore entries starting with .
-h, --human-readable    with -l and -s, print sizes like 1K 234M 2G etc.
-r, --reverse           reverse order while sorting...
-t                      sort by time, newest first; see --time...
```

> [!NOTE]
> from the feedback of a great guy u/stianhoiland i am writing a rewrite of `whats` in a bash script, the execution of this idea of a command explainer is fairly simple and writing it in a single bash script makes it more easier to run, and is a much more elegant solution than writing a whole go repo

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

---
<a href="https://www.star-history.com/?repos=iamkaran%2Fwhats&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=iamkaran/whats&type=date&theme=dark&legend=top-left&sealed_token=YeuiOHLARRNaxON9PX9ypCe7hXtNy0AwbULF6OgvRAz1EOMOK0EqnuDRpIQRhC_-GFAj0oMg06iLIJM_DM-tlxqi-FU6mWBVKy17lxxhhhnTsu7bOM3Ndg" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=iamkaran/whats&type=date&legend=top-left&sealed_token=YeuiOHLARRNaxON9PX9ypCe7hXtNy0AwbULF6OgvRAz1EOMOK0EqnuDRpIQRhC_-GFAj0oMg06iLIJM_DM-tlxqi-FU6mWBVKy17lxxhhhnTsu7bOM3Ndg" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=iamkaran/whats&type=date&legend=top-left&sealed_token=YeuiOHLARRNaxON9PX9ypCe7hXtNy0AwbULF6OgvRAz1EOMOK0EqnuDRpIQRhC_-GFAj0oMg06iLIJM_DM-tlxqi-FU6mWBVKy17lxxhhhnTsu7bOM3Ndg" />
 </picture>
</a>
