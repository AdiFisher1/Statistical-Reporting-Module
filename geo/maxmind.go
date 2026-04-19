package geo

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type MaxMindCountry struct {
	db *geoip2.Reader
}

// OpenMaxMindCountry opens the given .mmdb path (GeoLite2-Country.mmdb).
func OpenMaxMindCountry(path string) (*MaxMindCountry, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &MaxMindCountry{db: db}, nil
}

// Close releases the database handle.
func (m *MaxMindCountry) Close() error {
	return m.db.Close()
}

// Country returns the English country name when available, otherwise ISO 3166-1 alpha-2.
// Reserved/private IPs and missing data may return ("", err) or ("", nil) depending on the DB.
func (m *MaxMindCountry) Country(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP: %q", ipStr)
	}
	rec, err := m.db.Country(ip)
	if err != nil {
		return "", err
	}
	if n, ok := rec.Country.Names["en"]; ok && n != "" {
		return n, nil
	}
	if rec.Country.IsoCode != "" {
		return rec.Country.IsoCode, nil
	}
	return "", nil
}
