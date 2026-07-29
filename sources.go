package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

var (
	overStrike    = regexp.MustCompile(`.\x08`)
	flagLine      = regexp.MustCompile(`^(\s{1,8})(-.*?)(?:\s{2,}(.*))?$`)
	flagCluster   = regexp.MustCompile(`,\s+|\s+\|\s+`)
	subCmdLine    = regexp.MustCompile(`^([ \t]{1,8})(\S+?)(?:\s{2,}(.*))?$`)
	synopsisBlock = regexp.MustCompile(`(?s)SYNOPSIS\n(.*?)\n[A-Z][A-Z ]+\n`)
)

// source is a struct that represents a source of information like man page, --help output and many other
// it comes with a set of functions allowing the source to retain it's own way of capturing targets
type source interface {
	name() string
	fetch(args []string) (string, error)
	parseFlags(text string, args []string) []entry
	parseSubcmd(text string, args []string) []entry
}

// CmdManPage is a struct implementing source interface with functions for parsing and fetching man page
type CmdManPage struct{}

func (s *CmdManPage) name() string {
	return "command man page"
}

func (s *CmdManPage) fetch(args []string) (string, error) {
	return manPage(args[0])
}

func (s *CmdManPage) parseFlags(text string, args []string) []entry {
	if m := matchedEntries(args[1:], text); len(m) > 0 {
		return m
	}
	return []entry{}
}

func (s *CmdManPage) parseSubcmd(text string, args []string) []entry {
	subEntries := parseSubCmd(text)
	result := []entry{}
	for _, sc := range subEntries {
		for _, a := range args {
			if slices.Contains(sc.flags, a) {
				result = append(result, sc)
			}
		}
	}
	return result
}

// SubCmdManPage is a struct implementing source interface with functions for parsing and fetching man page
type SubCmdManPage struct{}

func (s *SubCmdManPage) name() string {
	return "sub command man page"
}

func (s *SubCmdManPage) fetch(args []string) (string, error) {
	if !strings.HasPrefix(args[1], "-") {
		return manPage(args[0] + args[1])
	} else {
		return manPage(args[0])
	}
}

func (s *SubCmdManPage) parseFlags(text string, args []string) []entry {
	if m := matchedEntries(args[1:], text); len(m) > 0 {
		return m
	}
	return []entry{}
}

func (s *SubCmdManPage) parseSubcmd(text string, args []string) []entry {
	// TODO: chain sub commands
	subEntries := parseSubCmd(text)
	result := []entry{}
	for _, sc := range subEntries {
		for _, a := range args {
			if slices.Contains(sc.flags, a) {
				result = append(result, sc)
			}
		}
	}
	return result
}

// synopsisCheck checks if the line after SYNOPSIS section show a genuine sub command (cmd sub [...) or a command (cmd-sub [...)
// a good example of this can be ssh-add where a command like ssh add (host named add) can trigger a false man page
func synopsisCheck(text, name string) bool {
	m := synopsisBlock.FindStringSubmatch(text)
	if m == nil {
		return false
	}
	word := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
	return word.MatchString(m[1])
}

// manPage returns manpage with formatting removed
func manPage(page string) (string, error) {
	c := exec.Command("man", page)
	c.Env = append(os.Environ(), "MANWIDTH=80", "TERM=DUMB")
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
