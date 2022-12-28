package netbox

import (
	"net/netip"

	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/models"
)

// Extracted from netbox/models/device's Device, but reduced to basic
// Go types to make it easier to work with the subset that I care
// about.
type Device struct {
	CustomFields       map[string]interface{}
	DeviceRole         string          // DeviceRole.Name
	DeviceType         string          // DeviceType.Model
	DeviceManufacturer string          // DeviceType.Manufacturer.Name
	Display            string          // Display
	ID                 int64           // ID
	Location           string          // Location.Name
	Name               string          // Name
	ParentDeviceID     int64           // ParentDevice.ID
	Platform           string          // Platform.Name
	PrimaryIP          netip.Prefix    // PrimaryIP.Address
	PrimaryIP4         netip.Prefix    // PrimaryIP4.Address
	PrimaryIP6         netip.Prefix    // PrimaryIP6.Address
	Tags               map[string]bool // Tags.Name -> true
}

type Devices []*Device

func dcimDeviceToDevice(dev *models.DeviceWithConfigContext) (*Device, error) {
	d := &Device{
		CustomFields: make(map[string]interface{}),
		Display:      dev.Display,
		ID:           dev.ID,
		Name:         *dev.Name,
		Tags:         make(map[string]bool),
	}

	//		if dev.CustomFields != nil {
	//		}
	if dev.DeviceRole != nil {
		d.DeviceRole = *dev.DeviceRole.Name
	}
	if dev.DeviceType != nil {
		d.DeviceType = *dev.DeviceType.Model
		if dev.DeviceType.Manufacturer != nil {
			d.DeviceType = *dev.DeviceType.Manufacturer.Name
		}
	}
	if dev.Location != nil {
		d.Location = *dev.Location.Name
	}
	if dev.ParentDevice != nil {
		d.ParentDeviceID = dev.ParentDevice.ID
	}
	if dev.Platform != nil {
		d.Platform = *dev.Platform.Name
	}
	if dev.PrimaryIP != nil {
		d.PrimaryIP = netip.MustParsePrefix(*dev.PrimaryIP.Address)
	}
	if dev.PrimaryIp4 != nil {
		d.PrimaryIP4 = netip.MustParsePrefix(*dev.PrimaryIp4.Address)
	}
	if dev.PrimaryIp6 != nil {
		d.PrimaryIP6 = netip.MustParsePrefix(*dev.PrimaryIp6.Address)
	}

	//		for _, cf := range results.custom_fields {
	//			d.CustomFields[
	//		}

	for _, t := range dev.Tags {
		d.Tags[*t.Name] = true
	}

	return d, nil
}

func ListDevices(c *client.NetBoxAPI) (Devices, error) {
	var limit int64
	limit = 0

	r := dcim.NewDcimDevicesListParams()
	r.Limit = &limit

	rs, err := c.Dcim.DcimDevicesList(r, nil)
	if err != nil {
		return nil, err
	}

	devs := make(Devices, len(rs.Payload.Results))
	for i, result := range rs.Payload.Results {
		d, err := dcimDeviceToDevice(result)
		if err != nil {
			return nil, err
		}

		devs[i] = d
	}

	return devs, nil
}

func GetDevice(c *client.NetBoxAPI, id int64) (*Device, error) {
	var limit int64
	limit = 0
	idStr := string(id)

	r := dcim.NewDcimDevicesListParams()
	r.Limit = &limit
	r.ID = &idStr

	rs, err := c.Dcim.DcimDevicesList(r, nil)
	if err != nil {
		return nil, err
	}

	return dcimDeviceToDevice(rs.Payload.Results[0])
}
