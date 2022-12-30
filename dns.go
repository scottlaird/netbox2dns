package netbox2dns

import (
	"context"
	"fmt"
)

type DNSProvider interface {
	ImportZone(cz *ConfigZone) (*Zone, error)
	WriteRecord(cz *ConfigZone, r *Record) error
	RemoveRecord(cz *ConfigZone, r *Record) error
	Save(cz *ConfigZone) error
}

func NewDNSProvider(ctx context.Context, cz *ConfigZone) (DNSProvider, error) {
	switch cz.ZoneType {
	case "clouddns":
		return NewCloudDNS(ctx, cz)
	case "zonefile":
		return NewZoneFileDNS(ctx, cz)
	default:
		return nil, fmt.Errorf("Unknown DNS provider type %q", cz.ZoneType)
	}
}

func ImportZones(ctx context.Context, cfg *Config) (*Zones, error) {
	zones := NewZones()

	for _, cz := range cfg.ZoneMap {
		provider, err := NewDNSProvider(ctx, cz)
		if err != nil {
			return nil, fmt.Errorf("Unable to get provider for zone %q: %v", cz.Name, err)
		}
		zone, err := provider.ImportZone(cz)
		if err != nil {
			return nil, fmt.Errorf("Unable to get import zone: %v", err)
		}

		zones.AddZone(zone)
	}
	return zones, nil
}

/*

// WriteRecord adds a single new record to the DNS provider.  We
// explicitly don't write entire zones, as they contain RRs that
// Netbox doesn't natively model (SOA and NS at a minimum).
func WriteRecord(ctx context.Context, cz *ConfigZone, r *Record) error {
	provider, err := NewDNSProvider(ctx, cz)
	if err != nil {
		return fmt.Errorf("Unable to get provider for zone %q: %v", cz.Name, err)
	}
	return provider.WriteRecord(cz, r)
}

// RemoveRecord removes a single record from the DNS provider.
// Attempting to remove a record that doesn't exist (has already been
// removed, etc) is not treated as an error.
func RemoveRecord(ctx context.Context, cz *ConfigZone, r *Record) error {
	provider, err := NewDNSProvider(ctx, cz)
	if err != nil {
		return fmt.Errorf("Unable to get provider for zone %q: %v", cz.Name, err)
	}
	return provider.RemoveRecord(cz, r)
}
*/
