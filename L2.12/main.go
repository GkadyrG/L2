package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	After       int      // -A N: lines after match
	Before      int      // -B N: lines before match
	Context     int      // -C N: lines around match
	Count       bool     // -c: count matches only
	IgnoreCase  bool     // -i: ignore case
	Invert      bool     // -v: invert match
	FixedString bool     // -F: fixed string match
	LineNumber  bool     // -n: show line numbers
	Pattern     string   // search pattern
	Files       []string // input files
}

type Match struct {
	LineNumber int
	Content    string
	IsMatch    bool // true if this line is an actual match, false if it's context
}

type Matcher interface {
	Match(line string) bool
}

type RegexMatcher struct {
	regex *regexp.Regexp
}

func (rm *RegexMatcher) Match(line string) bool {
	return rm.regex.MatchString(line)
}

type FixedMatcher struct {
	pattern    string
	ignoreCase bool
}

func (fm *FixedMatcher) Match(line string) bool {
	if fm.ignoreCase {
		return strings.Contains(strings.ToLower(line), strings.ToLower(fm.pattern))
	}
	return strings.Contains(line, fm.pattern)
}

func parseFlags() (*Config, error) {
	config := &Config{}

	flag.IntVar(&config.After, "A", 0, "print N lines after each match")
	flag.IntVar(&config.Before, "B", 0, "print N lines before each match")
	flag.IntVar(&config.Context, "C", 0, "print N lines of context around each match")
	flag.BoolVar(&config.Count, "c", false, "count matches only")
	flag.BoolVar(&config.IgnoreCase, "i", false, "ignore case")
	flag.BoolVar(&config.Invert, "v", false, "invert match")
	flag.BoolVar(&config.FixedString, "F", false, "interpret pattern as fixed string")
	flag.BoolVar(&config.LineNumber, "n", false, "show line numbers")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: grep [OPTIONS] PATTERN [FILE...]")
	}

	config.Pattern = args[0]
	config.Files = args[1:]

	// -C flag sets both -A and -B
	if config.Context > 0 {
		config.After = config.Context
		config.Before = config.Context
	}

	return config, nil
}

func createMatcher(config *Config) (Matcher, error) {
	if config.FixedString {
		return &FixedMatcher{
			pattern:    config.Pattern,
			ignoreCase: config.IgnoreCase,
		}, nil
	}

	pattern := config.Pattern
	if config.IgnoreCase {
		pattern = "(?i)" + pattern
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	return &RegexMatcher{regex: regex}, nil
}

func processReader(reader io.Reader, matcher Matcher, config *Config) ([]Match, error) {
	var lines []string
	var matches []Match

	scanner := bufio.NewScanner(reader)
	lineNum := 0

	// Read all lines first
	for scanner.Scan() {
		lineNum++
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %v", err)
	}

	// Find matches
	matchLines := make(map[int]bool)
	contextLines := make(map[int]bool)

	for i, line := range lines {
		lineNumber := i + 1
		isMatch := matcher.Match(line)

		// Apply invert logic
		if config.Invert {
			isMatch = !isMatch
		}

		if isMatch {
			matchLines[lineNumber] = true

			// Add context lines
			for j := max(1, lineNumber-config.Before); j <= min(len(lines), lineNumber+config.After); j++ {
				contextLines[j] = true
			}
		}
	}

	// Build result
	for lineNumber := 1; lineNumber <= len(lines); lineNumber++ {
		if contextLines[lineNumber] {
			match := Match{
				LineNumber: lineNumber,
				Content:    lines[lineNumber-1],
				IsMatch:    matchLines[lineNumber],
			}
			matches = append(matches, match)
		}
	}

	return matches, nil
}

func formatOutput(matches []Match, config *Config, filename string) {
	if config.Count {
		count := 0
		for _, match := range matches {
			if match.IsMatch {
				count++
			}
		}
		if filename != "" {
			fmt.Printf("%s:%d\n", filename, count)
		} else {
			fmt.Printf("%d\n", count)
		}
		return
	}

	printed := make(map[int]bool)

	for _, match := range matches {
		if printed[match.LineNumber] {
			continue
		}
		printed[match.LineNumber] = true

		var output strings.Builder

		if filename != "" {
			output.WriteString(filename)
			output.WriteString(":")
		}

		if config.LineNumber {
			output.WriteString(fmt.Sprintf("%d:", match.LineNumber))
		}

		output.WriteString(match.Content)
		fmt.Println(output.String())
	}
}

func processFile(filename string, matcher Matcher, config *Config) error {
	var reader io.Reader
	var file *os.File
	var err error

	if filename == "" || filename == "-" {
		reader = os.Stdin
		filename = ""
	} else {
		file, err = os.Open(filename)
		if err != nil {
			return fmt.Errorf("cannot open file %s: %v", filename, err)
		}
		defer file.Close()
		reader = file
	}

	matches, err := processReader(reader, matcher, config)
	if err != nil {
		return err
	}

	formatOutput(matches, config, filename)
	return nil
}

func main() {
	config, err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	matcher, err := createMatcher(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(config.Files) == 0 {
		config.Files = []string{"-"}
	}

	hasErrors := false
	for _, filename := range config.Files {
		if err := processFile(filename, matcher, config); err != nil {
			fmt.Fprintf(os.Stderr, "grep: %v\n", err)
			hasErrors = true
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
