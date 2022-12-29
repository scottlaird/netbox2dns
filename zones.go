package netbox2dns

import (
	"fmt"
	log "github.com/golang/glog"
	"github.com/scottlaird/netbox2dns/netbox"
	"net/netip"
	"sort"
	"strings"
)

// Type ByLength is a wrapper for []string for sorting the string
// slice by length, from longest to shortest.
type ByLength []*Zone

func (a ByLength) Len() int {
	return len(a)
}
func (a ByLength) Less(i, j int) bool {
	return len(a[i].Name) > len(a[j].Name)
}
func (a ByLength) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type Zones struct {
	Zones       map[string]*Zone
	sortedZones []*Zone
}

func NewZones() *Zones {
	return &Zones{
		Zones: make(map[string]*Zone),
	}
}

func (z *Zones) AddRecord(r *Record) error {
	for _, zone := range z.sortedZones {
		if strings.HasSuffix(r.Name, zone.Name+".") {
			zone.AddRecord(r)
			return nil
		}
	}
	return fmt.Errorf("Can't find zone matching record %q in %v", r.Name, z.sortedZones)
}

func (z *Zones) AddZone(zone *Zone) {
	z.Zones[zone.Name] = zone
	z.sortZones()
}

func (z *Zones) NewZone(cz *ConfigZone) {
	zone := Zone{
		Name:          cz.Name,
		ZoneName:      cz.ZoneName,
		Project:       cz.Project,
		Filename:      cz.Filename,
		DeleteEntries: cz.DeleteEntries,
		Ttl:           cz.Ttl,
		Records:       make(map[string][]*Record),
	}
	z.AddZone(&zone)
}

func (z *Zones) sortZones() {
	zones := make([]*Zone, len(z.Zones))
	i := 0
	for _, zone := range z.Zones {
		zones[i] = zone
		i++
	}
	sort.Sort(ByLength(zones))

	z.sortedZones = zones
}

func (z *Zones) Compare(newer *Zones) []*ZoneDelta {
	zones := make(map[string]bool)
	deltas := []*ZoneDelta{}

	// Create union of zones in z and newer
	for k := range z.Zones {
		zones[k] = true
	}
	for k := range newer.Zones {
		zones[k] = true
	}

	for k := range zones {
		if z.Zones[k] == nil {
			// Only in 'newer'
			fmt.Printf("*** Added Zone %q\n", k)
		} else if newer.Zones[k] == nil {
			// Only in 'z'
			fmt.Printf("*** Removed Zone %q\n", k)
		} else {
			zd := z.Zones[k].NewZoneDelta()
			z.Zones[k].Compare(newer.Zones[k], zd)
			deltas = append(deltas, zd)
		}
	}
	return deltas
}

type Zone struct {
	Name          string
	ZoneName      string
	Project       string
	Filename      string
	DeleteEntries bool
	Ttl           int64
	Records       map[string][]*Record
}

func (z *Zone) AddRecord(r *Record) {
	if r.Ttl == 0 {
		r.Ttl = z.Ttl
	}
	z.Records[r.Name] = append(z.Records[r.Name], r)
}

func (z *Zone) Compare(newer *Zone, zd *ZoneDelta) {
	records := make(map[string]bool)

	// Create union of zones in z and newer
	for k := range z.Records {
		records[k] = true
	}
	for k := range newer.Records {
		records[k] = true
	}

	for k := range records {
		if z.Records[k] == nil {
			// Only in 'newer'
			zd.AddRecords[k] = newer.Records[k]
		} else if newer.Records[k] == nil {
			// Only in 'z'
			if z.DeleteEntries {
				zd.RemoveRecords[k] = z.Records[k]
			}
		} else {
			CompareRecordSets(z.Records[k], newer.Records[k], zd)
		}
	}
}

func (z *Zone) NewZoneDelta() *ZoneDelta {
	zd := &ZoneDelta{
		Name:          z.Name,
		ZoneName:      z.ZoneName,
		Project:       z.Project,
		Filename:      z.Filename,
		AddRecords:    make(map[string][]*Record),
		RemoveRecords: make(map[string][]*Record),
	}
	return zd
}

type ZoneDelta struct {
	Name          string
	ZoneName      string
	Project       string
	Filename      string
	AddRecords    map[string][]*Record
	RemoveRecords map[string][]*Record
}

func CompareRecordSets(older []*Record, newer []*Record, zd *ZoneDelta) {
	// So, let's start by looking for identical Records.

	o := make([]string, len(older))
	n := make([]string, len(newer))

	for i, r := range older {
		o[i] = fmt.Sprintf("%+v", r)
	}

	for i, r := range newer {
		n[i] = fmt.Sprintf("%+v", r)
	}

	// Now, let's start by removing duplicates.  These sets should
	// be small, so O(N^2) is fine.
	for i, r := range o {
		for j, s := range n {
			if r == s {
				// Duplicate!  Remove from each set.
				o[i] = ""
				n[j] = ""
			}
		}
	}

	// At this point, any non-"" entries in o or n are actual deltas.
	for i, r := range o {
		if r != "" {
			name := older[i].Name
			zd.RemoveRecords[name] = append(zd.RemoveRecords[name], older[i])
		}
	}
	for i, r := range n {
		if r != "" {
			name := newer[i].Name
			zd.AddRecords[name] = append(zd.AddRecords[name], newer[i])
		}
	}
}

func ReverseName(addr netip.Addr) string {
	if addr.Is4() {
		return reverseName4(addr)
	}
	return reverseName6(addr)
}

func reverseName4(addr netip.Addr) string {
	b := addr.As4()
	return fmt.Sprintf("%d.%d.%d.%d.in-addr.arpa.", b[3], b[2], b[1], b[0])
}

func reverseName6(addr netip.Addr) string {
	ret := ""
	b := addr.As16()
	for i := 15; i >= 0; i-- {
		ret = ret + fmt.Sprintf("%x.%x.", b[i]&0xf, (b[i]&0xf0)>>4)
	}
	return ret + "ip6.arpa."
}

func (z *Zones) AddAddrs(addrs netbox.IPAddrs) error {
	for _, addr := range addrs {
		if addr.DNSName != "" && addr.Status == "active" {
			forward := Record{
				Name:    addr.DNSName + ".",
				Rrdatas: []string{addr.Address.Addr().String()},
			}
			reverse := Record{
				Name:    ReverseName(addr.Address.Addr()),
				Type:    "PTR",
				Rrdatas: []string{addr.DNSName + "."},
			}
			if addr.Address.Addr().Is4() {
				forward.Type = "A"
			} else {
				forward.Type = "AAAA"
			}

			err := z.AddRecord(&forward)
			if err != nil {
				return fmt.Errorf("Unable to add forward record for %q: %v", addr.DNSName, err)
			}
			err = z.AddRecord(&reverse)
			if err != nil {
				log.Warningf("Unable to add reverse record: %v", err)
			}
		}
	}
	return nil
}
