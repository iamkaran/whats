package main

import (
	"slices"
	"testing"
)

func TestFlagLine(t *testing.T) {
	tests := []struct {
		name      string
		lookup    string // the flag being queried, as in `whats <cmd> <flag>`
		rawBlock  string
		wantFlags []string
		wantDesc  string
	}{
		{
			name:   "ls --all: same-line description",
			lookup: "-a",
			rawBlock: `       -a, --all[=WHEN]        do not ignore entries starting with .
       -A, --almost-all       do not list implied . and ..
       -B, --ignore-backups   do not list entries ending with ~`,
			wantFlags: []string{"-a", "--all[=WHEN]"},
			wantDesc:  "do not ignore entries starting with .",
		},
		{
			name:   "sort --output=FILE: indented multi-line description",
			lookup: "-o",
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
			lookup: "--color",
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
			lookup: "-q",
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
			lookup: "-I",
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
			lookup: "--backup",
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
			lookup: "-o",
			rawBlock: `  -o FILE, --output FILE
                        write output to FILE

  -h, --help
                        show this help message and exit`,
			wantFlags: []string{"-o FILE", "--output FILE"},
			wantDesc:  "write output to FILE",
		},
		{
			name:   "pflag --port: lookup skips earlier flags in the block",
			lookup: "-p",
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
			lookup: "--log-level",
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
			lookup: "-j",
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
			lookup: "-H",
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
			lookup: "--dry-run",
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
			lookup: "--verify",
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
			lookup: "--example",
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
			lookup: "-#",
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
			lookup: "-NUM",
			rawBlock: `       -NUM   Same as --context=NUM.

       -A NUM, --after-context=NUM
               Print NUM lines of trailing context after matching lines.`,
			wantFlags: []string{"-NUM"},
			wantDesc:  "Same as --context=NUM.",
		},
		{
			name:   "tail --lines=[+]NUM: bracketed sign inside the argument",
			lookup: "-n",
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
			lookup: "--color",
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
			lookup: "--archive",
			rawBlock: `        --archive, -a            archive mode; equals -rlptgoD (no -H,-A,-X)
        --verbose, -v            increase verbosity`,
			wantFlags: []string{"--archive", "-a"},
			wantDesc:  "archive mode; equals -rlptgoD (no -H,-A,-X)",
		},
		{
			name:   "-V | --version: pipe-separated flags",
			lookup: "-V",
			rawBlock: `       -V | --version
              Print version and exit.

       -h | --help
              Print help.`,
			wantFlags: []string{"-V", "--version"},
			wantDesc:  "Print version and exit.",
		},
		{
			name:   "git --no-decorate: stacked flag lines sharing one description",
			lookup: "--no-decorate",
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
			lookup: "--decorate",
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
			lookup: "-o",
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
			lookup: "-i",
			rawBlock: `  -i INPUT [INPUT ...], --input INPUT [INPUT ...]
                        input files to process
  -h, --help            show this help message and exit`,
			wantFlags: []string{"-i INPUT [INPUT ...]", "--input INPUT [INPUT ...]"},
			wantDesc:  "input files to process",
		},
		{
			name:   "git --[no-]color: optional negation prefix",
			lookup: "--[no-]color",
			rawBlock: `       --[no-]color
               Show colored diff. The default is best effort.

       --word-diff[=<mode>]
               Show a word diff.`,
			wantFlags: []string{"--[no-]color"},
			wantDesc:  "Show colored diff. The default is best effort.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := parse(tt.rawBlock)
			wantedEntry := findEntry(entries, tt.lookup)
			if wantedEntry == nil {
				t.Fatalf("didn't find the flag %q", tt.lookup)
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
