package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	log "github.com/golang/glog"
	"github.com/netbox-community/go-netbox/netbox/client"
	nb "github.com/scottlaird/netbox2dns"
	"github.com/scottlaird/netbox2dns/netbox"
)

var (
	config = flag.String("config", "", "Path of a config file, with a .yaml, .json, or .cue extension")
)

func usage() {
	fmt.Printf("Usage: netbox2dns [--config=FILE] diff|push\n")
	os.Exit(1)
}

func main() {
	flag.Parse()
	args := flag.Args()
	push := false

	if len(args) != 1 {
		usage()
	}

	switch args[0] {
	case "push":
		push = true
	case "diff":
		// nothing
	default:
		usage()
	}

	var err error

	file := *config
	if file == "" {
		file, err = nb.FindConfig("netbox2dns")
		if err != nil {
			log.Fatal(err)
		}
	}

	cfg, err := nb.ParseConfig(file)
	if err != nil {
		log.Fatalf("Failed to parse config: %v")
	}

	log.Infof("Config read: %+v", cfg)

	ctx := context.Background()

	zones, err := nb.ImportZones(ctx, cfg)
	if err != nil {
		log.Fatalf("Unable to import existing zones: %v", err)
	}

	b, err := json.MarshalIndent(zones, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal: %v", err)
	}

	log.Infof("Found %d zones (%d bytes)", len(zones.Zones), len(b))

	newZones := nb.NewZones()
	for _, cz := range cfg.ZoneMap {
		newZones.NewZone(cz)
	}

	transport := httptransport.New(cfg.Netbox.Host, client.DefaultBasePath, []string{"https"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", "Token "+cfg.Netbox.Token)
	c := client.New(transport, nil)

	addrs, err := netbox.ListIPAddrs(c)
	if err != nil {
		log.Fatalf("Unable to fetch IP Addresses from Netbox: %v", err)
	}

	fmt.Printf("Found %d IP Addresses in %d zones\n", len(addrs), len(newZones.Zones))

	err = newZones.AddAddrs(addrs)
	if err != nil {
		log.Fatalf("Unable to add IP addresses: %v", err)
	}

	b2, err := json.MarshalIndent(newZones, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal: %v", err)
	}

	log.Infof("Created %d zones (%d bytes)", len(newZones.Zones), len(b2))

	zd := zones.Compare(newZones)

	removeCount := 0
	addCount := 0

	for _, zone := range zd {
		changed := false

		provider, err := nb.NewDNSProvider(ctx, cfg.ZoneMap[zone.Name])
		if err != nil {
			log.Fatalf("Failed to create DNS provider for %q: %v", zone.Name, err)
		}

		for _, rec := range zone.RemoveRecords {
			for _, rr := range rec {
				if rr.Type == "A" || rr.Type == "AAAA" || rr.Type == "PTR" {
					removeCount++
					fmt.Printf("- %s %s %d %v\n", rr.Name, rr.Type, rr.Ttl, rr.Rrdatas)
					if push {
						err := provider.RemoveRecord(cfg.ZoneMap[zone.Name], rr)
						changed = true
						if err != nil {
							log.Errorf("Failed to remove record: %v", err)
						}
					}
				}
			}
		}
		for _, rec := range zone.AddRecords {
			for _, rr := range rec {
				addCount++
				fmt.Printf("+ %s %s %d %v\n", rr.Name, rr.Type, rr.Ttl, rr.Rrdatas)
				if push {
					err = provider.WriteRecord(cfg.ZoneMap[zone.Name], rr)
					changed = true
					if err != nil {
						log.Errorf("Failed to update record: %v", err)
					}
				}
			}
		}

		if changed {
			err := provider.Save(cfg.ZoneMap[zone.Name])
			if err != nil {
				log.Fatalf("Failed to save: %v", err)
			}
		}
	}

	if push {
		fmt.Printf("Push complete.  %d removals, %d additions found\n", removeCount, addCount)
	} else {
		fmt.Printf("Diff complete.  %d removals, %d additions found\n", removeCount, addCount)
	}
}
