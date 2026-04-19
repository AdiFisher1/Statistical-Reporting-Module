package report

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"project/parser"
)

// How many distinct values to list per dimension before grouping the rest as "Other".
const ReportTopN = 5

// Write prints frequency percentages for Country, OS, and Browser (each: top N + Other).
func Write(w io.Writer, entries []parser.Entry) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, "no entries")
		return err
	}
	total := float64(len(entries))

	sections := []struct {
		title string
		key   func(parser.Entry) string
	}{
		{"Country", func(e parser.Entry) string { return label(e.Country) }},
		{"OS", func(e parser.Entry) string { return label(e.OS) }},
		{"Browser", func(e parser.Entry) string { return label(e.Browser) }},
	}
	for i, sec := range sections {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := writeTopNOtherSection(w, sec.title, entries, total, sec.key, ReportTopN); err != nil {
			return err
		}
	}
	return nil
}

func label(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "Unknown"
	}
	return s
}

func countBy(entries []parser.Entry, key func(parser.Entry) string) map[string]int {
	m := make(map[string]int)
	for _, e := range entries {
		m[key(e)]++
	}
	return m
}

type pair struct {
	name  string
	count int
}

func sortedPairs(m map[string]int) []pair {
	out := make([]pair, 0, len(m))
	for k, v := range m {
		out = append(out, pair{k, v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].name < out[j].name
	})
	return out
}

func pct(count int, total float64) float64 {
	return 100.0 * float64(count) / total
}

func writeTopNOtherSection(w io.Writer, title string, entries []parser.Entry, total float64, key func(parser.Entry) string, topN int) error {
	m := countBy(entries, key)
	pairs := sortedPairs(m)
	if _, err := fmt.Fprintf(w, "%s:\n", title); err != nil {
		return err
	}
	for i := 0; i < len(pairs) && i < topN; i++ {
		p := pairs[i]
		if _, err := fmt.Fprintf(w, "%s %.2f%%\n", p.name, pct(p.count, total)); err != nil {
			return err
		}
	}
	other := 0
	for i := topN; i < len(pairs); i++ {
		other += pairs[i].count
	}
	if other > 0 {
		_, err := fmt.Fprintf(w, "Other %.2f%%\n", pct(other, total))
		return err
	}
	return nil
}
