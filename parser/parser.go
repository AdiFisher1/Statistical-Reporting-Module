package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	ua "github.com/mssola/user_agent"
)

type Entry struct {
	IP      string
	Country string
	OS      string
	Browser string
}

type LineParser interface {
	Parse(line string) (Entry, error)
}

// ApacheCombined implements LineParser for Apache combined log format
type ApacheCombined struct {
	re *regexp.Regexp
}

// Byte size may be "-" (no body / not applicable). Lines may be JSON-wrapped with \" for quotes.
var apacheCombinedPattern = regexp.MustCompile(
	`^(\S+) \S+ \S+ \[([^\]]+)\] "([^"]*)" (\d+) (\d+|-) "([^"]*)" "([^"]*)"$`,
)

// apacheCombinedLoose matches when the user-agent string is truncated (no closing ")
var apacheCombinedLoose = regexp.MustCompile(
	`^(\S+) \S+ \S+ \[([^\]]+)\] "([^"]*)" (\d+) (\d+|-) "([^"]*)" "(.*)$`,
)

func NewApacheCombined() *ApacheCombined {
	return &ApacheCombined{re: apacheCombinedPattern}
}

var (
	ErrEmptyLine   = errors.New("empty log line")
	ErrNoMatch     = errors.New("line does not match Apache combined log format")
	ErrNoUserAgent = errors.New("user-agent is empty")
)

// Mac OS X + numeric/underscore build → Mac OS
var reMacOSXNumeric = regexp.MustCompile(`(?i)\bMac OS X\b(?:\s+[\d._]+)+`)

// Longest first so "Intel Mac OS" wins over "Mac OS".
var osFamilyPrefixes = []string{
	"Intel Mac OS", "Mac OS", "Windows", "CrOS", "Ubuntu", "Android",
	"FreeBSD", "NetBSD", "OpenBSD", "SunOS", "Linux",
}

// simplifyOS drops version / arch tails (e.g. "Intel Mac OS X 10_9_1" → "intel mac os").
func simplifyOS(os string) string {
	os = strings.TrimSpace(reMacOSXNumeric.ReplaceAllString(os, "Mac OS"))
	os = strings.TrimSpace(os)
	if os == "" {
		return os
	}
	lower := strings.ToLower(os)
	for _, fam := range osFamilyPrefixes {
		sub := strings.ToLower(fam)
		if idx := strings.Index(lower, sub); idx >= 0 {
			os = strings.TrimSpace(os[:idx+len(fam)])
			break
		}
	}
	return strings.ToLower(os)
}

// normalizeApacheLine handles logs exported as JSON strings: whole line in quotes with \" inside.
func normalizeApacheLine(line string) string {
	line = strings.TrimSpace(line)
	if n := len(line); n >= 2 && line[0] == '"' && line[n-1] == '"' {
		line = line[1 : n-1]
		line = strings.ReplaceAll(line, `\"`, `"`)
	}
	return line
}

// Parse extracts IP from the line and OS + browser from the User-Agent field.
func (p *ApacheCombined) Parse(line string) (Entry, error) {
	line = normalizeApacheLine(line)
	if line == "" {
		return Entry{}, ErrEmptyLine
	}
	m := p.re.FindStringSubmatch(line)
	if m == nil {
		m = apacheCombinedLoose.FindStringSubmatch(line)
	}
	if m == nil {
		return Entry{}, fmt.Errorf("%w: %q", ErrNoMatch, truncate(line, 120))
	}
	ip := m[1]
	userAgent := strings.TrimSuffix(m[7], `"`)
	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		return Entry{IP: ip}, ErrNoUserAgent
	}
	client := ua.New(userAgent)
	browserName, _ := client.Browser()
	browser := strings.TrimSpace(browserName)
	osInfo := simplifyOS(client.OS())
	return Entry{
		IP:      ip,
		OS:      osInfo,
		Browser: browser,
	}, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
