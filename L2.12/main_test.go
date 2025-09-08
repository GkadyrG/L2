package main

import (
	"strings"
	"testing"
)

func TestFixedMatcher(t *testing.T) {
	m := &FixedMatcher{pattern: "foo", ignoreCase: false}
	if !m.Match("hello foo world") {
		t.Error("expected to match substring 'foo'")
	}
	if m.Match("bar") {
		t.Error("did not expect match")
	}
}

func TestFixedMatcherIgnoreCase(t *testing.T) {
	m := &FixedMatcher{pattern: "foo", ignoreCase: true}
	if !m.Match("HELLO FOO WORLD") {
		t.Error("expected to match ignoring case")
	}
}

func TestRegexMatcher(t *testing.T) {
	rm, _ := createMatcher(&Config{Pattern: "f.o"})
	if !rm.Match("foo") {
		t.Error("expected regex to match foo")
	}
	if rm.Match("f123o") {
		t.Error("did not expect regex to match f123o")
	}
}

func TestSimpleMatch(t *testing.T) {
	input := "one\nfoo\nthree\n"
	config := &Config{Pattern: "foo"}
	matcher, _ := createMatcher(config)

	matches, err := processReader(strings.NewReader(input), matcher, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if !matches[0].IsMatch || matches[0].Content != "foo" {
		t.Error("expected to match line 'foo'")
	}
}

func TestInvertMatch(t *testing.T) {
	input := "one\nfoo\nthree\n"
	config := &Config{Pattern: "foo", Invert: true}
	matcher, _ := createMatcher(config)

	matches, _ := processReader(strings.NewReader(input), matcher, config)

	if len(matches) != 2 {
		t.Fatalf("expected 2 lines (invert), got %d", len(matches))
	}
	for _, m := range matches {
		if strings.Contains(m.Content, "foo") {
			t.Error("invert match should not contain 'foo'")
		}
	}
}

func TestContextBeforeAfter(t *testing.T) {
	input := "a\nb\nfoo\nd\ne\n"
	config := &Config{Pattern: "foo", Before: 1, After: 1}
	matcher, _ := createMatcher(config)

	matches, _ := processReader(strings.NewReader(input), matcher, config)

	if len(matches) != 3 {
		t.Fatalf("expected 3 lines (before+match+after), got %d", len(matches))
	}

	expected := []string{"b", "foo", "d"}
	for i, m := range matches {
		if m.Content != expected[i] {
			t.Errorf("expected %q, got %q", expected[i], m.Content)
		}
	}
}

func TestContextCFlag(t *testing.T) {
	input := "a\nb\nfoo\nd\ne\n"
	config := &Config{Pattern: "foo", Context: 2, Before: 2, After: 2}
	matcher, _ := createMatcher(config)

	matches, _ := processReader(strings.NewReader(input), matcher, config)

	// Should capture: a, b, foo, d, e
	if len(matches) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(matches))
	}
}

func TestCountFlag(t *testing.T) {
	input := "foo\nbar\nfoo\nbaz\n"
	config := &Config{Pattern: "foo", Count: true}
	matcher, _ := createMatcher(config)

	matches, _ := processReader(strings.NewReader(input), matcher, config)

	count := 0
	for _, m := range matches {
		if m.IsMatch {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 matches, got %d", count)
	}
}

func TestIgnoreCase(t *testing.T) {
	input := "FOO\nbar\n"
	config := &Config{Pattern: "foo", IgnoreCase: true}
	matcher, _ := createMatcher(config)

	matches, _ := processReader(strings.NewReader(input), matcher, config)

	if len(matches) != 1 {
		t.Fatalf("expected 1 match ignoring case, got %d", len(matches))
	}
	if matches[0].Content != "FOO" {
		t.Errorf("expected 'FOO', got %q", matches[0].Content)
	}
}

func TestMultipleMatchesWithOverlap(t *testing.T) {
	input := "a\nfoo\nb\nfoo\nc\n"
	config := &Config{Pattern: "foo", Before: 1, After: 1}
	matcher, _ := createMatcher(config)

	matches, _ := processReader(strings.NewReader(input), matcher, config)

	// Expect to see all lines since contexts overlap
	if len(matches) != 5 {
		t.Fatalf("expected 5 lines due to overlapping contexts, got %d", len(matches))
	}
}
