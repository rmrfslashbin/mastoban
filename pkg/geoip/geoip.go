package geoip

import (
	"errors"
	"net"
	"path/filepath"

	"github.com/oschwald/maxminddb-golang"
)

// Record defines the fields to fetch from the GeoIP database.
type Record struct {
	Continent struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"continent"`

	Country struct {
		IsoCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

// GeoIPData represents the data returned.
type GeoIPData struct {
	IP        net.IP `json:"ip"`
	Continent string `json:"continent_code"`
	Country   string `json:"country_code"`
}

// GeoIP struct is used to configure and exec geoip functions
type GeoIP struct {
	db *maxminddb.Reader
}

// New sets up and configures a new GeoIP struct for use
func New(dbfile *string) (*GeoIP, error) {
	// If the database file is not specified, use the default.
	dbfqpn := ""

	if dbfile != nil {
		// If the database file is specified, resolve it to an absolute path.
		if dbfqpnAbs, err := filepath.Abs(*dbfile); err != nil {
			return nil, errors.New("filepath.Abs returned an error parsing database file path '" + *dbfile + "': " + err.Error())
		} else {
			dbfqpn = dbfqpnAbs
		}
	}

	// Connect to the local MaxMind GeoIP database file
	if db, err := maxminddb.Open(dbfqpn); err != nil {
		return nil, errors.New("maxminddb.Open returned an error opening the database file '" + dbfqpn + "': " + err.Error())
	} else {
		return &GeoIP{db: db}, nil
	}
}

// Lookip GeoIP data for the given IP address.
func (geoIP *GeoIP) Lookup(ip net.IP) (*GeoIPData, error) {
	var data Record = Record{}

	err := geoIP.db.Lookup(ip, &data)
	if err != nil {
		return nil, errors.New("geoIP.db.Lookup returned an error looking up the IP address: " + err.Error())
	}

	// Return the IP addr, continent, and country data
	return &GeoIPData{
		IP:        ip,
		Continent: data.Continent.Code,
		Country:   data.Country.IsoCode,
	}, nil
}
