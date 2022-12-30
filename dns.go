package netbox2dns

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/golang/glog"
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

// function IncrementSerial increments the serial number on a DNS
// zone.  This recognizes 2 basic serial number patterns:
//
// 1.  Simple incrememnting integers (1 -> 2 -> 3, etc)
// 2.  Date-based serial numbers (2022123004 -> 2022123005 -> 2022123100)
//
// It assumes that any serial number greather than 2000010100 (the
// first date-based serial number in 2000 AD) is date-based and
// updates it accordingly.  If the date portion of the serial matches
// todays date, then it increments the serial number by one.  If the
// date-based portion does *not* match today's date, then the new
// serial number is today, with two trailing 0s.
//
// Finally, there is a check that the new serial is greater than the
// old serial.  This will break after 100 updates happen on a single
// day (2022123099 will get incremented to 2022123100, which is fine,
// but the following update will try to use 2022123000, which will
// fail).  This is a fundimental problem with date-based serial
// formats, and will clear up on its own once the calendar rolls over
// to the next day.
func IncrementSerial(cz *ConfigZone, serial uint32) (uint32, error) {
	today := time.Now().Format("20060102")
	return incrementSerialFixedDate(cz, serial, today)

}
func incrementSerialFixedDate(cz *ConfigZone, serial uint32, today string) (uint32, error) {
	if serial >= 2000_01_01_00 {
		log.Infof("Using date-based serial number for zone %q", cz.Name)
		// Using YYYYMMDDxx serial numbers, probably.
		today := time.Now().Format("20060102")

		serialString := strconv.FormatUint(uint64(serial), 10)

		// Have we already written a serial number today?  If
		// so, just increment.  This will break with >100
		// updates per day.
		if strings.HasPrefix(serialString, today) {
			serial++
			log.Infof("Updating date-based serial number by 1 to %d", serial)
		} else {
			log.Infof("Starting new date; %q is not a prefix of %q", today, serialString)
			newserial, _ := strconv.ParseUint(today, 10, 32)
			newserial *= 100
			if uint32(newserial) <= serial {
				return 0, fmt.Errorf("Can't figure out serial format; current value of %d is greater than proposed new value of %d", serial, newserial)
			}
			serial = uint32(newserial)
		}
	} else {
		serial++
	}

	return serial, nil
}
