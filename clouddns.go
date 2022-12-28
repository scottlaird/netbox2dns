package netbox2dns

import (
	"context"
	"fmt"

	"google.golang.org/api/dns/v1"
	//log "github.com/golang/glog"
)

type CloudDNS struct {
	rrss *dns.ResourceRecordSetsService
}

func NewCloudDNS(ctx context.Context) (*CloudDNS, error) {
	cd := &CloudDNS{}

	dnsService, err := dns.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to DNS: %v", err)
	}
	cd.rrss = dns.NewResourceRecordSetsService(dnsService)

	return cd, nil
}

func (cd *CloudDNS) ImportZone(cfg *ConfigZone) (*Zone, error) {
	zone := &Zone{
		Name:          cfg.Name,
		ZoneName:      cfg.ZoneName,
		Project:       cfg.Project,
		Ttl:           cfg.Ttl,
		DeleteEntries: cfg.DeleteEntries,
		Records:       make(map[string][]*Record),
	}

	call := cd.rrss.List(zone.Project, cfg.ZoneName)
	rrs, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to get zone: %v", err)
	}

	for _, r := range rrs.Rrsets {
		rr := Record{
			Name:    r.Name,
			Type:    r.Type,
			Ttl:     r.Ttl,
			Rrdatas: r.Rrdatas,
		}
		zone.AddRecord(&rr)
	}

	return zone, nil
}

func (cd *CloudDNS) WriteRecord(cz *ConfigZone, r *Record) error {
	c := cd.rrss.Create(cz.Project, cz.ZoneName, &dns.ResourceRecordSet{
		Name:    r.Name,
		Type:    r.Type,
		Ttl:     r.Ttl,
		Rrdatas: r.Rrdatas,
	})
	_, err := c.Do()
	return err
}

func (cd *CloudDNS) RemoveRecord(cz *ConfigZone, r *Record) error {
	fmt.Printf("(should) remove %s %s %d %v\n", r.Name, r.Type, r.Ttl, r.Rrdatas)
	return nil
}
