package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// TODO: Make this easily configurable
const maxLen = 80

// flagLine matches a flag entry: indented line starting with -x or
// --long, with
// the description either after 2+ spaces on the same line or on following lines.
var (
	overStrike    = regexp.MustCompile(`.\x08`)
	flagLine      = regexp.MustCompile(`^(\s{1,8})(-.*?)(?:\s{2,}(.*))?$`)
	flagCluster   = regexp.MustCompile(`,\s+|\s+\|\s+`)
	subCmdLine    = regexp.MustCompile(`^([ \t]{1,8})(\S+?)(?:\s{2,}(.*))?$`)
	synopsisBlock = regexp.MustCompile(`(?s)SYNOPSIS\n(.*?)\n[A-Z][A-Z ]+\n`)

	isDump = *flag.Bool("dump", false, "dump the entire parsed slice of entries")
)

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
	fmt.Println(args)

	var result []entry

	isSubCmd := false
	if !strings.HasPrefix(args[1], "-") {
		isSubCmd = true
	}

	c := []entry{}
	s := []entry{}

	if text, err := manPage(args[0]); err == nil && len(text) != 0 {
		var m []entry
		if isSubCmd {
			m = matchedEntries(args[2:], text)
		} else {
			m = matchedEntries(args[1:], text)
		}
		if len(m) > 0 {
			c = append(c, m...)
		}
	}

	var subManText string

	if isSubCmd {
		if text, err := manPage(args[0] + "-" + args[1]); err == nil && len(text) != 0 {
			subManText = text
			matches := matchedEntries(args[1:], text)
			if len(matches) > 0 {
				s = append(s, matches...)
			}

			subEntries := parseSubCmd(text)
			for _, sc := range subEntries {
				if slices.Contains(sc.flags, args[1]) {
					s = append(s, sc)
				}
			}
		}
	}

	genuineSubCmd := isSubCmd && subManText != "" && hasSubcommand(subManText, args[0]+"-"+args[1])

	if genuineSubCmd && len(c) > 0 {
		result = append(result, c...)
	} else if len(c) > len(s) {
		result = append(result, c...)
	} else {
		result = append(result, s...)
	}

	if len(result) < len(args[1:]) {
		if text, err := helpOutput(args[0], "--help"); err == nil && len(text) > 0 {
			var r []entry
			if isSubCmd {
				r = matchedEntries(args[2:], text)
			} else {
				r = matchedEntries(args[1:], text)
			}

			if len(r) > 0 {
				for _, e := range r {
					if !matchesArg(e, args) {
						continue
					}

					if containsEntry(result, e) {
						continue
					}

					result = append(result, e)
				}
			}
			if isSubCmd {
				subEntries := parseSubCmd(text)

				if len(subEntries) > 0 {
					for _, sc := range subEntries {
						if slices.Contains(sc.flags, args[1]) {
							if !containsEntry(result, sc) {
								result = append(result, sc)
							}
						}
					}
				}
			}
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

func hasSubcommand(text, name string) bool {
	m := synopsisBlock.FindStringSubmatch(text)
	if m == nil {
		return false
	}
	word := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
	return word.MatchString(m[1])
}

func entriesEqual(a, b entry) bool {
	if a.desc != b.desc {
		return false
	}

	if len(a.flags) != len(b.flags) {
		return false
	}

	for i := range a.flags {
		if a.flags[i] != b.flags[i] { // TODO: use reflect instead
			return false
		}
	}

	return true
}

func containsEntry(l []entry, e entry) bool {
	for _, existing := range l {
		if entriesEqual(existing, e) {
			return true
		}
	}
	return false
}

func matchesArg(e entry, args []string) bool {
	for _, f := range e.flags {
		for _, a := range args {
			if f == a {
				return true
			}
		}
	}

	return false
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

// manPage returns manpage with formatting removed
func manPage(page string) (string, error) {
	c := exec.Command("man", page)
	c.Env = append(os.Environ(), "MANWIDTH=80")
	out, err := c.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(overStrike.ReplaceAll(out, nil)), nil
}

// helpOutput returns the output of the --help flag or just the help as a sub command
func helpOutput(cmd, flag string) (string, error) {
	out, err := exec.Command(cmd, flag).CombinedOutput()
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		return "", fmt.Errorf("lenght of help output is zero")
	}
	return string(out), nil
}
