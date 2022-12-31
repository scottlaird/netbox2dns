package netbox2dns

import (
	"fmt"
	"strings"
)

// Record describes a DNS record, like 'foo.example.com IN AAAA 1:2::3:4'.
type Record struct {
	Name    string
	Type    string
	TTL     int64
	Rrdatas []string
}

// Compare compares two Records and prints a difference.
//
// TODO: is this used still?
func (r *Record) Compare(newer *Record) string {
	// Not comparing 'Name', because it's a key in Zone.Record, so
	// we shouldn't ever be called with mismatches.

	if r.Type != newer.Type {
		fmt.Printf("*** %s: Changed type from %q to %q\n", r.Name, r.Type, newer.Type)
	}
	if r.TTL != newer.TTL {
		fmt.Printf("*** %s: Changed ttl from %d to %d\n", r.Name, r.TTL, newer.TTL)
	}
	if r.Rrdatas[0] != newer.Rrdatas[0] {
		fmt.Printf("*** %s: Changed ttl from %v to %v\n", r.Name, r.Rrdatas, newer.Rrdatas)
	}

	return ""
}

// NameNoDot returns the name of a record with no trailing dot.
func (r *Record) NameNoDot() string {
	return strings.TrimRight(r.Name, ".")
}

// RrdataNoDot returns the Rrdata for a record, with no trailing dor.
func (r *Record) RrdataNoDot() string {
	return strings.TrimRight(r.Rrdatas[0], ".")
}
