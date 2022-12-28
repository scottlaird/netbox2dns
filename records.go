package netbox2dns

import (
	"fmt"
)

type Record struct {
	Name    string
	Type    string
	Ttl     int64
	Rrdatas []string
}

func (r *Record) Compare(newer *Record) string {
	// Not comparing 'Name', because it's a key in Zone.Record, so
	// we shouldn't ever be called with mismatches.

	if r.Type != newer.Type {
		fmt.Printf("*** %s: Changed type from %q to %q\n", r.Name, r.Type, newer.Type)
	}
	if r.Ttl != newer.Ttl {
		fmt.Printf("*** %s: Changed ttl from %d to %d\n", r.Name, r.Ttl, newer.Ttl)
	}
	if r.Rrdatas[0] != newer.Rrdatas[0] {
		fmt.Printf("*** %s: Changed ttl from %v to %v\n", r.Name, r.Rrdatas, newer.Rrdatas)
	}

	return ""
}
