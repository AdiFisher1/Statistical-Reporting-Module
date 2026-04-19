package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"project/geo"
	"project/parser"
	"project/report"
	"project/store"
)

func main() {
	path := flag.String("file", "", "path to Apache combined log file (required)")
	mmdb := flag.String("mmdb", "", "path to GeoLite2-Country.mmdb")
	flag.Parse()
	if *path == "" {
		flag.Usage()
		os.Exit(2)
	}

	f, err := os.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var lookup geo.CountryLookup
	if *mmdb != "" {
		mm, err := geo.OpenMaxMindCountry(*mmdb)
		if err != nil {
			log.Fatalf("open mmdb: %v", err)
		}
		defer mm.Close()
		lookup = mm
	}

	st := store.NewMemoryStore()
	if _, err := processLog(f, parser.NewApacheCombined(), lookup, st); err != nil {
		log.Fatal(err)
	}
	if err := report.Write(os.Stdout, st.Snapshot()); err != nil {
		log.Fatal(err)
	}
}

func processLog(r io.Reader, p parser.LineParser, lookup geo.CountryLookup, s store.EntryStore) (saved int, err error) {
	sc := bufio.NewScanner(r)
	// Default buffer may be too small for very long lines
	const max = 1024 * 1024
	buf := make([]byte, max)
	sc.Buffer(buf, max)

	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		entry, err := p.Parse(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "line %d: %v\n", lineNo, err)
			continue
		}
		if lookup != nil {
			c, err := lookup.Country(entry.IP)
			if err != nil {
				fmt.Fprintf(os.Stderr, "line %d: country lookup: %v\n", lineNo, err)
			} else {
				entry.Country = c
			}
		}
		if err := s.Save(entry); err != nil {
			return saved, fmt.Errorf("line %d: save: %w", lineNo, err)
		}
		saved++
	}
	return saved, sc.Err()
}
