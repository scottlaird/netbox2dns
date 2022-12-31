package netbox2dns

import (
	"strings"
)

// Record describes a DNS record, like 'foo.example.com IN AAAA 1:2::3:4'.
type Record struct {
	Name    string
	Type    string
	TTL     int64
	Rrdatas []string
}

// NameNoDot returns the name of a record with no trailing dot.
func (r *Record) NameNoDot() string {
	return strings.TrimRight(r.Name, ".")
}

// RrdataNoDot returns the Rrdata for a record, with no trailing dor.
func (r *Record) RrdataNoDot() string {
	return strings.TrimRight(r.Rrdatas[0], ".")
}
