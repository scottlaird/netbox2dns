package netbox2dns

import (
	"context"
	"fmt"

	"google.golang.org/api/dns/v1"
)

// TODO: starting using dns.Changes to bundle up multiple changes into
// a single transaction.  This should substantailly improve
// performance when making multiple changes at once.

// CloudDNS implements talking to Google Cloud DNS, and provides
// methods for fetching existing DNS entries, adding new entries, or
// deleting old entries.
type CloudDNS struct {
	rrss *dns.ResourceRecordSetsService
}

// NewCloudDNS creates a new CloudDNS.
func NewCloudDNS(ctx context.Context, cz *ConfigZone) (*CloudDNS, error) {
	cd := &CloudDNS{}

	dnsService, err := dns.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to DNS: %v", err)
	}
	cd.rrss = dns.NewResourceRecordSetsService(dnsService)

	return cd, nil
}

// ImportZone imports all entries from the specified Google Cloud DNS zone.
func (cd *CloudDNS) ImportZone(cfg *ConfigZone) (*Zone, error) {
	zone := &Zone{
		Name:          cfg.Name,
		ZoneName:      cfg.ZoneName,
		Project:       cfg.Project,
		TTL:           cfg.TTL,
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
			TTL:     r.Ttl,
			Rrdatas: r.Rrdatas,
		}
		zone.AddRecord(&rr)
	}

	return zone, nil
}

// WriteRecord adds a record to Google Cloud DNS.
func (cd *CloudDNS) WriteRecord(cz *ConfigZone, r *Record) error {
	c := cd.rrss.Create(cz.Project, cz.ZoneName, &dns.ResourceRecordSet{
		Name:    r.Name,
		Type:    r.Type,
		Ttl:     r.TTL,
		Rrdatas: r.Rrdatas,
	})
	_, err := c.Do()
	return err
}

// RemoveRecord removes a DNS entry from Google Cloud DNS.
//
// TODO: implement
func (cd *CloudDNS) RemoveRecord(cz *ConfigZone, r *Record) error {
	d := cd.rrss.Delete(cz.Project, cz.ZoneName, r.Name, r.Type)
	_, err := d.Do()

	return err
}

// Save flushes changes to Cloud DNS.  This is a no-op at the moment,
// but we'll eventually batch queries together for performance.
func (cd *CloudDNS) Save(cz *ConfigZone) error {
	return nil
}

