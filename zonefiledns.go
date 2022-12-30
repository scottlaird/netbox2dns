package netbox2dns

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/golang/glog"
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
func (zfd *ZoneFileDNS) IncrementSerial(cz *ConfigZone) error {
	serial := zfd.zone.SOA.Serial
	if serial > 2000_01_01_00 {
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
				return fmt.Errorf("Can't figure out serial format; current value of %d is greater than proposed new value of %d", serial, newserial)
			}
			serial = uint32(newserial)
		}
	} else {
		serial++
	}

	zfd.zone.SOA.Serial = serial

	return nil
}

func (zfd *ZoneFileDNS) Save(cz *ConfigZone) error {
	err := zfd.IncrementSerial(cz)
	if err != nil {
		return err
	}

	return zfd.zone.Save()
}
