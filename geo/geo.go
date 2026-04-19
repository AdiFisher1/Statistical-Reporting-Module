package geo

type CountryLookup interface {
	Country(ip string) (country string, err error)
}
