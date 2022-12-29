package netbox2dns

import (
	"context"
	"strings"

	"github.com/shuLhan/share/lib/dns"
	//"github.com/bwesterb/go-zonefile"
	//log "github.com/golang/glog"
)

type ZoneFileDNS struct {
}

func NewZoneFileDNS(ctx context.Context) (*ZoneFileDNS, error) {
	cd := &ZoneFileDNS{}

	return cd, nil
}

func (cd *ZoneFileDNS) ImportZone(cz *ConfigZone) (*Zone, error) {
	zone := &Zone{
		Name:          cz.Name,
		Filename:      cz.Filename,
		Ttl:           cz.Ttl,
		DeleteEntries: cz.DeleteEntries,
		Records:       make(map[string][]*Record),
	}

	zf, err := dns.ParseZoneFile(zone.Filename, zone.Name, uint32(zone.Ttl))
	if err != nil {
		return nil, err
	}

	for _, i := range zf.Records {
		for _, entry := range i {
			r := Record{
				Name: strings.TrimRight(entry.Name, ".") + ".",
				Type: dns.RecordTypeNames[entry.Type],
				Ttl:  int64(entry.TTL),
			}
			
			s := entry.Value.(string)
			if entry.Type ==  dns.RecordTypePTR {
				// TODO verify trailing .
			}
			r.Rrdatas = []string{strings.TrimRight(s, ".") + "."}

			zone.AddRecord(&r)
		}
	}

	return zone, nil
}

func (cd *ZoneFileDNS) rrFromRecord(cz *ConfigZone, r *Record) *dns.ResourceRecord {
	return &dns.ResourceRecord{
		Name: r.NameNoDot(),
		Type: dns.RecordTypes[r.Type],
		Class: dns.RecordClassIN,
		TTL: uint32(r.Ttl),
		Value: r.RrdataNoDot(),
	}
}

func (cd *ZoneFileDNS) WriteRecord(cz *ConfigZone, r *Record) error {
	zf, err := dns.ParseZoneFile(cz.Filename, cz.Name, uint32(cz.Ttl))
	if err != nil {
		return err
	}

	entry := cd.rrFromRecord(cz, r)

	err = zf.Add(entry)
	if err != nil {
		return err
	}

	return zf.Save()
}

func (cd *ZoneFileDNS) RemoveRecord(cz *ConfigZone, r *Record) error {
	zf, err := dns.ParseZoneFile(cz.Filename, cz.Name, uint32(cz.Ttl))
	if err != nil {
		return err
	}

	entry := cd.rrFromRecord(cz, r)

	err = zf.Remove(entry)
	if err != nil {
		return err
	}

	return zf.Save()
}
