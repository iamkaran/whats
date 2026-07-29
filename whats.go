package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
)

// TODO: Make this easily configurable
const maxLen = 80

// flagLine matches a flag entry: indented line starting with -x or
// --long, with
// the description either after 2+ spaces on the same line or on following lines.
var isDump = *flag.Bool("dump", false, "dump the entire parsed slice of entries")

// entry is a flag and its description
type entry struct {
	flags []string
	desc  string
}

func init() {
	flag.Usage = func() {
		_, err := fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		if err != nil {
			panic(err)
		}
		_, err = fmt.Fprintln(flag.CommandLine.Output(), "  <cmd> <subcommand> <flags>\n    	Print the explanation of the command")
		if err != nil {
			panic(err)
		}
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	args := flag.Args()
	if args[0] == "sudo" {
		args = slices.Delete(args, 0, 1)
	}

	sources := []source{
		&CmdManPage{},
		&SubCmdManPage{},
	}

	result := []entry{}

	for _, s := range sources {
		text, err := s.fetch(args)
		if err != nil {
			continue
		}

		cmdMatches := s.parseFlags(text, args)
		for _, m := range cmdMatches {
			if isDuplicate(result, m) {
				continue
			}
			result = append(result, m)
		}
		subCmdMatches := s.parseSubcmd(text, args)
		for _, m := range subCmdMatches {
			if isDuplicate(result, m) {
				continue
			}
			result = append(result, m)
		}
	}

	w := 0
	joined := make([]string, len(result))
	for i, en := range result {
		joined[i] = strings.Join(en.flags, ", ")
		if n := len([]rune(joined[i])); n > w {
			w = n
		}
	}
	for i, en := range result {
		pad := strings.Repeat(" ", w-len([]rune(joined[i])))
		fmt.Printf("%s%s    %s\n", strings.Join(en.flags, ", "), pad, getSummary(en.desc))
	}
}

// isDuplicate checks if slice A has any element of entry B's slice of Flags
func isDuplicate(result []entry, m entry) bool {
	return slices.ContainsFunc(result, func(r entry) bool {
		return slices.ContainsFunc(r.flags, func(a string) bool {
			return slices.Contains(m.flags, a)
		})
	})
}

// matchedEntries returns a slice of entry's that match flags in args
// reason for abstraction: it handles combined args as well
func matchedEntries(args []string, text string) []entry {
	entries := parseFlags(text)
	if len(entries) == 0 || isDump {
		return entries
	}

	result := []entry{}
	for _, arg := range args {
		matchedEntry := findEntry(entries, flagName(arg))
		// Combined flags
		if matchedEntry == nil && strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 2 {
			combinedFlags := strings.TrimPrefix(arg, "-")
			fannedOut := strings.Split(combinedFlags, "")
			for _, fArg := range fannedOut {
				matchedEntry := findEntry(entries, "-"+flagName(fArg))
				if matchedEntry != nil {
					result = append(result, *matchedEntry)
				}
			}
		} else {
			if matchedEntry != nil {
				result = append(result, *matchedEntry)
			}
		}
	}
	return result
}

// parse returns a slice of entries containing flags and their description from the source
func parseSubCmd(text string) []entry {
	lines := strings.Split(text, "\n")
	var entries []entry
	for i, line := range lines {
		if m := subCmdLine.FindStringSubmatch(line); m != nil {
			if strings.HasPrefix(m[2], "-") {
				continue
			}
			description := descAfter(lines, i, m[3])
			if description == "" {
				description = descAfter(lines, i+1, m[3])
			}
			entries = append(entries, entry{flags: []string{m[2]}, desc: description})
		}
	}
	return entries
}

// parseFlags returns a slice of entries containing flags and their description from the source
func parseFlags(text string) []entry {
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
	if i := strings.Index(f, "["); i != -1 {
		f = f[:i]
	}
	if i := strings.Index(f, "[="); i != -1 {
		f = f[:i]
	}
	if i := strings.IndexAny(f, " ="); i != -1 {
		f = f[:i]
	}
	if i := strings.IndexAny(f, "("); i != -1 {
		f = f[:1]
	}
	return f
}

// getSummary returns the first line from the entire description block
func getSummary(desc string) string {
	const term = ".:)]}!?"
	for _, ln := range strings.Split(desc, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		r := []rune(ln)
		if !strings.ContainsRune(term, r[len(r)-1]) {
			ln += "..."
		}
		if r := []rune(ln); len(r) > maxLen {
			ln = string(r[:maxLen-1]) + "…"
		}
		return ln
	}
	return ""
}
