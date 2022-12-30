package netbox2dns

import (
	"context"
	"strings"

	"github.com/shuLhan/share/lib/dns"
)

type ZoneFileDNS struct {
	zone *dns.Zone
}

func NewZoneFileDNS(ctx context.Context, cz *ConfigZone) (*ZoneFileDNS, error) {
	zone, err := dns.ParseZoneFile(cz.Filename, cz.Name, uint32(cz.Ttl))
	if err != nil {
		return nil, err
	}

	zfd := &ZoneFileDNS{
		zone: zone,
	}

	return zfd, nil
}

func (zfd *ZoneFileDNS) ImportZone(cz *ConfigZone) (*Zone, error) {
	zone := &Zone{
		Name:          cz.Name,
		Filename:      cz.Filename,
		Ttl:           cz.Ttl,
		DeleteEntries: cz.DeleteEntries,
		Records:       make(map[string][]*Record),
	}

	for _, i := range zfd.zone.Records {
		for _, entry := range i {
			r := Record{
				Name: strings.TrimRight(entry.Name, ".") + ".",
				Type: dns.RecordTypeNames[entry.Type],
				Ttl:  int64(entry.TTL),
			}

			s := entry.Value.(string)
			if entry.Type == dns.RecordTypePTR {
				// TODO verify trailing .
			}
			r.Rrdatas = []string{strings.TrimRight(s, ".") + "."}

			zone.AddRecord(&r)
		}
	}

	return zone, nil
}

func (zfd *ZoneFileDNS) rrFromRecord(cz *ConfigZone, r *Record) *dns.ResourceRecord {
	return &dns.ResourceRecord{
		Name:  r.NameNoDot(),
		Type:  dns.RecordTypes[r.Type],
		Class: dns.RecordClassIN,
		TTL:   uint32(r.Ttl),
		Value: r.RrdataNoDot(),
	}
}

func (zfd *ZoneFileDNS) WriteRecord(cz *ConfigZone, r *Record) error {
	entry := zfd.rrFromRecord(cz, r)

	err := zfd.zone.Add(entry)
	if err != nil {
		return err
	}

	return nil
}

func (zfd *ZoneFileDNS) RemoveRecord(cz *ConfigZone, r *Record) error {
	entry := zfd.rrFromRecord(cz, r)

	err := zfd.zone.Remove(entry)
	if err != nil {
		return err
	}
	return nil
}

func (zfd *ZoneFileDNS) Save(cz *ConfigZone) error {
	newserial, err := IncrementSerial(cz, zfd.zone.SOA.Serial)
	if err != nil {
		return err
	}
	zfd.zone.SOA.Serial = newserial

	return zfd.zone.Save()
}
