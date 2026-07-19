package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

// flagLine matches a flag entry: indented line starting with -x or
// --long, possibly a comma or pipe separated additional flags ("-l, --long[=WHEN]"), with
// the description either after 2+ spaces on the same line or on following lines.
var (
	flagLine = regexp.MustCompile(
		`^(?: {1,12})` + // Indent at the start of a line
			`(-{1,2}[^\s,=\[|]+)` + // The first flag
			`((?:\s?(?:[,|]\s{0,2})?` + // Additional flag separator (', ' or ' | ')
			`-{1,2}[^\s,=\[]+)*)` + // Additional flags
			`((?:\s|\[?=)\S+)?` + // Optional argument (" FILE", "=ARG", or "[=ARG]")
			`(?:\s{2,}(.*))?$`, // Description separated with 2 whitespaces
	)
	overstrike = regexp.MustCompile(`.\x08`)
)

// render returns manpage with formatting removed
func render(page string) (string, error) {
	cmd := exec.Command("man", "1", page)
	cmd.Env = append(os.Environ(), "MANWIDTH=80")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no man page for %q", page)
	}
	return string(overstrike.ReplaceAll(out, nil)), nil
}
