package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	httptransport "github.com/go-openapi/runtime/client"
	log "github.com/golang/glog"
	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/scottlaird/netbox2dns/netbox"
	nb "github.com/scottlaird/netbox2dns"
)

var (
	dryRun       = flag.Bool("dry_run", true, "Actually do things")
)

func main() {
	flag.Parse()

	file, err := nb.FindConfig("netbox2dns")
	if err != nil {
		log.Fatal(err)
	}
	
	
	cfg, err := nb.ParseConfig(file)
	if err != nil {
		log.Fatalf("Failed to parse config: %v")
	}

	fmt.Printf("Config: %+v\n", cfg)

	ctx := context.Background()
	
	zones, err := nb.ImportZones(ctx, cfg)
	if err != nil {
		log.Fatalf("Unable to import existing zones: %v", err)
	}

	b, err := json.MarshalIndent(zones, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal: %v", err)
	}

	fmt.Printf("Found %d zones (%d bytes)\n", len(zones.Zones), len(b))

	newZones := nb.NewZones()
	for _, cz := range cfg.ZoneMap {
		newZones.NewZone(cz)
	}

	transport := httptransport.New(cfg.Netbox.Host, client.DefaultBasePath, []string{"https"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+cfg.Netbox.Token)
	c := client.New(transport, nil)

	addrs, err := netbox.ListIPAddrs(c)
	for _, addr := range addrs {
		if addr.DNSName != "" && addr.Status == "active" {
			forward := nb.Record{
				Name:    addr.DNSName + ".",
				Rrdatas: []string{addr.Address.Addr().String()},
			}
			reverse := nb.Record{
				Name:    nb.ReverseName(addr.Address.Addr()),
				Type:    "PTR",
				Rrdatas: []string{addr.DNSName + "."},
			}
			if addr.Address.Addr().Is4() {
				forward.Type = "A"
			} else {
				forward.Type = "AAAA"
			}

			err = newZones.AddRecord(&forward)
			if err != nil {
				log.Fatalf("Unable to add forward record: %v", err)
			}
			err = newZones.AddRecord(&reverse)
			if err != nil {
				log.Warningf("Unable to add reverse record: %v", err)
			}
		}
	}
	b2, err := json.MarshalIndent(newZones, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal: %v", err)
	}

	//	fmt.Printf("JSON: %s\n", string(b2))
	fmt.Printf("Created %d zones (%d bytes)\n", len(newZones.Zones), len(b2))

	zd := zones.Compare(newZones)

	for _, zone := range zd {
		for _, rec := range zone.RemoveRecords {
			for _, rr := range rec {
				if rr.Type == "A" || rr.Type == "AAAA" || rr.Type == "PTR" {
					fmt.Printf("- %s %s %d %v\n", rr.Name, rr.Type, rr.Ttl, rr.Rrdatas)
					if !*dryRun {
						err := nb.RemoveRecord(ctx, cfg.ZoneMap[zone.Name], rr)
						if err != nil {
							log.Errorf("Failed to remove record: %v", err)
						}
					}
				}
			}
		}
		for _, rec := range zone.AddRecords {
			for _, rr := range rec {
				fmt.Printf("+ %s %s %d %v\n", rr.Name, rr.Type, rr.Ttl, rr.Rrdatas)
				if !*dryRun {
					err = nb.WriteRecord(ctx, cfg.ZoneMap[zone.Name], rr)
					if err != nil {
						log.Errorf("Failed to update record: %v", err)
					}
				}
			}
		}
	}
}
