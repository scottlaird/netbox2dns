package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	log "github.com/golang/glog"
	nb "github.com/scottlaird/netbox2dns"
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

	// Load config file
	file := *config
	if file == "" {
		file, err = nb.FindConfig("netbox2dns")
		if err != nil {
			log.Fatal(err)
		}
	}
	cfg, err := nb.ParseConfig(file)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	log.Infof("Config read: %+v", cfg)

	ctx := context.Background()

	// Fetch existing DNS zones and entries
	zones, err := nb.ImportZones(ctx, cfg)
	if err != nil {
		log.Fatalf("Unable to import existing zones: %v", err)
	}

	log.Infof("Found %d zones", len(zones.Zones))

	// Create new zones using data from Netbox
	newZones := nb.NewZones()
	for _, cz := range cfg.ZoneMap {
		newZones.NewZone(cz)
	}

	addrs, err := nb.GetNetboxIPAddresses(cfg.Netbox.Host, cfg.Netbox.Token)
	if err != nil {
		log.Fatalf("Unable to fetch IP Addresses from Netbox: %v", err)
	}

	fmt.Printf("Found %d IP Addresses in %d zones\n", len(addrs), len(newZones.Zones))

	// Add Netbox IPs to our new zones
	err = newZones.AddAddrs(addrs)
	if err != nil {
		log.Fatalf("Unable to add IP addresses: %v", err)
	}

	log.Infof("Created %d zones", len(newZones.Zones))

	// Compare imported zones to created zones and produce a diff.
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
					fmt.Printf("- %s %s %d %v\n", rr.Name, rr.Type, rr.TTL, rr.Rrdatas)
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
				fmt.Printf("+ %s %s %d %v\n", rr.Name, rr.Type, rr.TTL, rr.Rrdatas)
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
