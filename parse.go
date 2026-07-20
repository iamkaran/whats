package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// flagLine matches a flag entry: indented line starting with -x or
// --long, with
// the description either after 2+ spaces on the same line or on following lines.
var (
	flagLine    = regexp.MustCompile(`^( {1,8})(-.*?)(?:\s{2,}(.*))?$`)
	flagCluster = regexp.MustCompile(`,\s+|\s+\|\s+`)

	overstrike = regexp.MustCompile(`.\x08`)
)

// entry is a flag and its description
type entry struct {
	flags []string
	desc  string
}

// checkDump returns the arguments with removed --dump flag if there is one and a bool indicating if the user wants to dump all entries
func checkDump(args []string) ([]string, bool) {
	if len(args) >= 1 && args[0] == "--dump" {
		return os.Args[2:], true
	}
	return args, false
}

func main() {
	if len(os.Args) <= 2 {
		fmt.Fprintf(os.Stderr, "usage: whats <cmd> <flags>\n")
		os.Exit(1)
	}

	args, dump := checkDump(os.Args[1:])

	text, err := render(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "no man page found\n")
		os.Exit(1)
	}
	entries := parse(text)
	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "no entries found\n")
		os.Exit(1)
	}

	if dump == true {
		for _, en := range entries {
			fmt.Printf("%s    %s\n", strings.Join(en.flags, ", "), en.desc)
		}
		os.Exit(0)
	}
	result := []*entry{}
	for _, arg := range args {
		entry := findEntry(entries, flagName(arg))
		if entry != nil {
			result = append(result, entry)
		}
	}
	if len(result) == 0 {
		fmt.Fprintf(os.Stderr, "no flag matched\n")
		os.Exit(1)
	}
	for _, en := range result {
		fmt.Printf("%s    %s\n", strings.Join(en.flags, ", "), en.desc)
	}
}

// render returns manpage with formatting removed
func render(page string) (string, error) {
	cmd := exec.Command("man", page)
	cmd.Env = append(os.Environ(), "MANWIDTH=80")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no man page for %q", page)
	}
	return string(overstrike.ReplaceAll(out, nil)), nil
}

// parse returns a slice of entries containing flags and their description from the source
func parse(text string) []entry {
	lines := strings.Split(text, "\n")
	var entries []entry
lineLoop:
	for i, line := range lines {
		if m := flagLine.FindStringSubmatch(line); m != nil {
			flags := flagCluster.Split(m[2], -1)
			for _, flag := range flags {
				if !strings.HasPrefix(flag, "-") {
					continue lineLoop
				}
			}
			description := descAfter(lines, i, m[3])
			// stacked synonymed (eg: git --no-decorate) have their description in next flag
			if description == "" {
				description = descAfter(lines, i+1, m[3])
			}
			entries = append(entries, entry{flags: flags, desc: description})
		}
	}
	return entries
}

// descAfter returns the entire description string following a flagline
func descAfter(lines []string, start int, seed string) string {
	descIndent := -1
	desc := seed
	blanks := 0
	for j := start + 1; j < len(lines); j++ {
		line := lines[j]
		if flagLine.MatchString(line) {
			break
		}

		if line == "" {
			blanks++
			continue
		}

		if descIndent == -1 {
			descIndent = leadingIndent(line)
		}

		if leadingIndent(line) < descIndent {
			break
		}

		// Preserve new lines
		if desc != "" {
			desc += strings.Repeat("\n", blanks+1)
		}
		blanks = 0
		desc += strings.TrimPrefix(line, strings.Repeat(" ", descIndent))
	}
	return desc
}

// leadingIndent returns the amount of spaces the start of the lines by counting columns
func leadingIndent(s string) int {
	cols := 0
	for _, r := range s {
		switch r {
		case ' ':
			cols++
		case '\t':
			cols += 8 - cols%8
		default:
			return cols
		}
	}
	return len(s) - len(strings.TrimLeft(s, " \t"))
}

// findEntry returns the entry object for a flag from the slice of entries
func findEntry(entries []entry, flag string) *entry {
	for i := range entries {
		for _, f := range entries[i].flags {
			if flagName(f) == flag {
				return &entries[i]
			}
		}
	}
	return nil
}

// flagName strips the argument from a stored flag: "-o option",
// "--color[=WHEN]", "--output=FILE" all reduce to their bare flag.
func flagName(f string) string {
	if i := strings.Index(f, "[="); i != -1 {
		f = f[:i]
	}
	if i := strings.IndexAny(f, " ="); i != -1 {
		f = f[:i]
	}
	return f
}
