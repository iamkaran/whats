package main

import (
	"slices"
	"testing"
)

func TestFlagLine(t *testing.T) {
	tests := []struct {
		name      string
		lookUp    string // the flag being queried, as in `whats <cmd> <flag>`
		rawBlock  string
		wantFlags []string
		wantDesc  string
	}{
		{
			name:   "ls --all: same-line description",
			lookUp: "-a",
			rawBlock: `       -a, --all[=WHEN]        do not ignore entries starting with .
       -A, --almost-all       do not list implied . and ..
       -B, --ignore-backups   do not list entries ending with ~`,
			wantFlags: []string{"-a", "--all[=WHEN]"},
			wantDesc:  "do not ignore entries starting with .",
		},
		{
			name:   "sort --output=FILE: indented multi-line description",
			lookUp: "-o",
			rawBlock: `       -o, --output=FILE
               Write output to FILE instead of standard output.
               Existing files are truncated before writing.
               Use "-" to write to stdout.

       -a, --append
               Append to FILE instead of truncating.`,
			wantFlags: []string{"-o", "--output=FILE"},
			wantDesc: `Write output to FILE instead of standard output.
Existing files are truncated before writing.
Use "-" to write to stdout.`,
		},
		{
			name:   "grep --color[=WHEN]: description with blank line and value list",
			lookUp: "--color",
			rawBlock: `       --color[=WHEN]
               Colorize the output.

               WHEN may be:
                 always
                 auto
                 never

       --help
               Display this help and exit.`,
			wantFlags: []string{"--color[=WHEN]"},
			wantDesc: `Colorize the output.

WHEN may be:
  always
  auto
  never`,
		},
		{
			name:   "grep --quiet: four synonym flags",
			lookUp: "-q",
			rawBlock: `       -q, -s, --quiet, --silent
               Suppress all normal output.
               Errors are still reported.

       -v, --verbose
               Produce more detailed output.`,
			wantFlags: []string{"-q", "-s", "--quiet", "--silent"},
			wantDesc: `Suppress all normal output.
Errors are still reported.`,
		},
		{
			name:   "gcc --include=DIR: space-separated arg on the short flag",
			lookUp: "-I",
			rawBlock: `       -I DIR, --include=DIR
               Add DIR to the include search path.

               This option may be specified multiple times.
               Directories are searched in the order given.

       -D NAME=VALUE
               Define a macro.`,
			wantFlags: []string{"-I DIR", "--include=DIR"},
			wantDesc: `Add DIR to the include search path.

This option may be specified multiple times.
Directories are searched in the order given.`,
		},
		{
			name:   "cp --backup[=CONTROL]: same-line description plus indented list",
			lookUp: "--backup",
			rawBlock: `       --backup[=CONTROL]        make a backup of each existing destination file

               CONTROL may be one of:
                 none
                 numbered
                 existing
                 simple

       --suffix=SUFFIX
               override the usual backup suffix`,
			wantFlags: []string{"--backup[=CONTROL]"},
			wantDesc: `make a backup of each existing destination file

CONTROL may be one of:
  none
  numbered
  existing
  simple`,
		},
		{
			name:   "argparse --output FILE: space-separated args on both flags",
			lookUp: "-o",
			rawBlock: `  -o FILE, --output FILE
                        write output to FILE

  -h, --help
                        show this help message and exit`,
			wantFlags: []string{"-o FILE", "--output FILE"},
			wantDesc:  "write output to FILE",
		},
		{
			name:   "pflag --port: lookUp skips earlier flags in the block",
			lookUp: "-p",
			rawBlock: `Flags:
      --config string      path to config file
  -p, --port int           TCP port to listen on
      --tls                enable TLS
  -h, --help               help for server`,
			wantFlags: []string{"-p", "--port int"},
			wantDesc:  "TCP port to listen on",
		},
		{
			name:   "argparse --log-level: braced choices as argument",
			lookUp: "--log-level",
			rawBlock: `Options:
  --log-level {debug,info,warning,error}
                        logging verbosity

  --format {json,yaml,toml}
                        output serialization format

  -h, --help
                        show this help message and exit`,
			wantFlags: []string{"--log-level {debug,info,warning,error}"},
			wantDesc:  "logging verbosity",
		},
		{
			name:   "make --jobs=N: same-line description with continuation lines",
			lookUp: "-j",
			rawBlock: `       -j, --jobs=N         Run up to N jobs simultaneously.
                            Defaults to the number of online CPUs.
                            Values less than one are rejected.

       --load-average=N    Do not start new jobs if load is too high.`,
			wantFlags: []string{"-j", "--jobs=N"},
			wantDesc: `Run up to N jobs simultaneously.
Defaults to the number of online CPUs.
Values less than one are rejected.`,
		},
		{
			name:   "curl --header: flag-like examples inside the description",
			lookUp: "-H",
			rawBlock: `       -H, --header HEADER
               Add HEADER to every HTTP request.

               Example:
                 -H "Authorization: Bearer <token>"
                 -H "Accept: application/json"

       -d, --data DATA
               HTTP request body`,
			wantFlags: []string{"-H", "--header HEADER"},
			wantDesc: `Add HEADER to every HTTP request.

Example:
  -H "Authorization: Bearer <token>"
  -H "Accept: application/json"`,
		},
		{
			name:   "--dry-run: quoted flags inside the description",
			lookUp: "--dry-run",
			rawBlock: `       --dry-run
               Does not execute commands.
               For example, "-f" and "--force" are ignored
               when running in dry-run mode.

       --force
               Force execution.`,
			wantFlags: []string{"--dry-run"},
			wantDesc: `Does not execute commands.
For example, "-f" and "--force" are ignored
when running in dry-run mode.`,
		},
		{
			name:   "--verify: bulleted list in the description",
			lookUp: "--verify",
			rawBlock: `       --verify
               Verification stages:
                 - Parse the configuration
                 - Validate signatures
                 - Compare hashes
               Returns non-zero on failure.

       --repair
               Attempt automatic repair.`,
			wantFlags: []string{"--verify"},
			wantDesc: `Verification stages:
  - Parse the configuration
  - Validate signatures
  - Compare hashes
Returns non-zero on failure.`,
		},
		{
			name:   "--example: command invocations inside the description",
			lookUp: "--example",
			rawBlock: `       --example
               Example:
                 mytool -a --output=file
                 mytool --color=always

       --help
               Show help.`,
			wantFlags: []string{"--example"},
			wantDesc: `Example:
  mytool -a --output=file
  mytool --color=always`,
		},
		{
			name:   "curl --progress-bar: '#' as short flag",
			lookUp: "-#",
			rawBlock: `       -#, --progress-bar
               Make curl display transfer progress as a simple progress bar
               instead of the standard, more informational, meter.

       -0, --http1.0
               (HTTP) Tells curl to use HTTP version 1.0`,
			wantFlags: []string{"-#", "--progress-bar"},
			wantDesc: `Make curl display transfer progress as a simple progress bar
instead of the standard, more informational, meter.`,
		},
		{
			name:   "grep -NUM: numeric context flag with same-line description",
			lookUp: "-NUM",
			rawBlock: `       -NUM   Same as --context=NUM.

       -A NUM, --after-context=NUM
               Print NUM lines of trailing context after matching lines.`,
			wantFlags: []string{"-NUM"},
			wantDesc:  "Same as --context=NUM.",
		},
		{
			name:   "tail --lines=[+]NUM: bracketed sign inside the argument",
			lookUp: "-n",
			rawBlock: `       -n, --lines=[+]NUM
               output the last NUM lines, instead of the last 10;
               or use -n +NUM to output starting with line NUM

       -q, --quiet, --silent
               never output headers giving file names`,
			wantFlags: []string{"-n", "--lines=[+]NUM"},
			wantDesc: `output the last NUM lines, instead of the last 10;
or use -n +NUM to output starting with line NUM`,
		},
		{
			name:   "grep --colour: alias long flags each carrying [=WHEN]",
			lookUp: "--color",
			rawBlock: `       --color[=WHEN], --colour[=WHEN]
               Surround the matched strings with escape sequences to display
               them in color on the terminal.

       --binary-files=TYPE
               Assume that binary files are TYPE.`,
			wantFlags: []string{"--color[=WHEN]", "--colour[=WHEN]"},
			wantDesc: `Surround the matched strings with escape sequences to display
them in color on the terminal.`,
		},
		{
			name:   "rsync --archive: long flag listed before short flag",
			lookUp: "--archive",
			rawBlock: `        --archive, -a            archive mode; equals -rlptgoD (no -H,-A,-X)
        --verbose, -v            increase verbosity`,
			wantFlags: []string{"--archive", "-a"},
			wantDesc:  "archive mode; equals -rlptgoD (no -H,-A,-X)",
		},
		{
			name:   "-V | --version: pipe-separated flags",
			lookUp: "-V",
			rawBlock: `       -V | --version
              Print version and exit.

       -h | --help
              Print help.`,
			wantFlags: []string{"-V", "--version"},
			wantDesc:  "Print version and exit.",
		},
		{
			name:   "git --no-decorate: stacked flag lines sharing one description",
			lookUp: "--no-decorate",
			rawBlock: `       --no-decorate
       --decorate[=short|full|auto|no]
               Print out the ref names of any commits that are shown. If short is
               specified, the ref name prefixes refs/heads/, refs/tags/ and
               refs/remotes/ will not be printed.

       --source
               Print out the ref name given on the command line by which each
               commit was reached.`,
			wantFlags: []string{"--no-decorate"},
			wantDesc: `Print out the ref names of any commits that are shown. If short is
specified, the ref name prefixes refs/heads/, refs/tags/ and
refs/remotes/ will not be printed.`,
		},
		{
			name:   "git --decorate: pipe-separated values inside the argument",
			lookUp: "--decorate",
			rawBlock: `       --no-decorate
       --decorate[=short|full|auto|no]
               Print out the ref names of any commits that are shown.

       --source
               Print out the ref name given on the command line.`,
			wantFlags: []string{"--decorate[=short|full|auto|no]"},
			wantDesc:  "Print out the ref names of any commits that are shown.",
		},
		{
			name:   "ssh -o ssh_option: BSD man indentation, entry at end of text",
			lookUp: "-o",
			rawBlock: `     -o ssh_option
             Can be used to give options in the format used in the
             configuration file. This is useful for specifying options for
             which there is no separate command-line flag.`,
			wantFlags: []string{"-o ssh_option"},
			wantDesc: `Can be used to give options in the format used in the
configuration file. This is useful for specifying options for
which there is no separate command-line flag.`,
		},
		{
			name:   "argparse --input: nargs metavars with brackets and spaces",
			lookUp: "-i",
			rawBlock: `  -i INPUT [INPUT ...], --input INPUT [INPUT ...]
                        input files to process
  -h, --help            show this help message and exit`,
			wantFlags: []string{"-i INPUT [INPUT ...]", "--input INPUT [INPUT ...]"},
			wantDesc:  "input files to process",
		},
		{
			name:   "git --[no-]color: optional negation prefix",
			lookUp: "--[no-]color",
			rawBlock: `       --[no-]color
               Show colored diff. The default is best effort.

       --word-diff[=<mode>]
               Show a word diff.`,
			wantFlags: []string{"--[no-]color"},
			wantDesc:  "Show colored diff. The default is best effort.",
		},
		{
			name:      "git log --min-parents=<number> false positive",
			lookUp:    "--min-parents=2",
			wantFlags: []string{"--min-parents=<number>", "--max-parents=<number>", "--no-min-parents,"},
			wantDesc: `Show only commits which have at least (or at most) that many parent
commits. In particular, --max-parents=1 is the same as --no-merges,
--min-parents=2 is the same as --merges.  --max-parents=0 gives all
root commits and --min-parents=3 all octopus merges.

--no-min-parents and --no-max-parents reset these limits (to no
limit) again. Equivalent forms are --min-parents=0 (any commit has
0 or more parents) and --max-parents=-1 (negative numbers denote no
upper limit).`,
			rawBlock: `
       --merges
           Print only merge commits. This is exactly the same as
           --min-parents=2.

       --no-merges
           Do not print commits with more than one parent. This is exactly the
           same as --max-parents=1.

       --min-parents=<number>, --max-parents=<number>, --no-min-parents,
       --no-max-parents
           Show only commits which have at least (or at most) that many parent
           commits. In particular, --max-parents=1 is the same as --no-merges,
           --min-parents=2 is the same as --merges.  --max-parents=0 gives all
           root commits and --min-parents=3 all octopus merges.

           --no-min-parents and --no-max-parents reset these limits (to no
           limit) again. Equivalent forms are --min-parents=0 (any commit has
           0 or more parents) and --max-parents=-1 (negative numbers denote no
           upper limit).
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := parse(tt.rawBlock)
			lookUp := flagName(tt.lookUp)
			wantedEntry := findEntry(entries, lookUp)
			if wantedEntry == nil {
				t.Logf("raw entries: %v\n", entries)
				t.Fatalf("didn't find the flag %q", lookUp)
			}
			if !slices.Equal(tt.wantFlags, wantedEntry.flags) {
				t.Fatalf("wanted %v got %v", tt.wantFlags, wantedEntry.flags)
			}
			if tt.wantDesc != wantedEntry.desc {
				t.Fatalf("wanted %q got %q", tt.wantDesc, wantedEntry.desc)
			}
		})
	}
}
