package netbox

import (
	"fmt"

	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/dcim"
	"github.com/netbox-community/go-netbox/netbox/models"
)

// Extracted from netbox/models/device's Interface, but reduced to
// basic Go types to make it easier to work with the subset that I
// care about.
type Interface struct {
	CableID                     int64
	CableDisplay                string
	CableLabel                  string
	CableEnd                    string
	ConnectedEndpoints          []string
	ConnectedEndpointsReachable bool // why is it a bool when ConnectedEndpoints is a list?
	CustomFields                map[string]interface{}
	Description                 string
	DeviceID                    int64
	Display                     string
	Enabled                     bool
	ID                          int64
	Label                       string
	LAGName                     string // lag.name
	LinkPeers                   []string
	LinkPeersType               string
	MACAddress                  string
	MarkConnected               bool
	MgmtOnly                    bool
	MTU                         int64
	Name                        string
	ParentID                    int64
	ParentName                  string
	PoeMode                     string // poe_mode.value
	PoeType                     string // poe_type.value
	Speed                       int64
	TaggedVLANNames             []string
	TaggedVLANIDs               []int64
	TaggedVLANVIDs              []int64
	Tags                        map[string]bool // Tags.Name -> true
	Type                        string          // type.value
	UntaggedVLANName            string
	UntaggedVLANID              int64
	UntaggedVLANVID             int64
	VRF                         string   // presumably?  Not currently using.
	WirelessLANs                []string // wireless_lans.ssid

	// Defined in NetBox 3.3 but not populated in my config.
	//
	//  Bridge
	//  ConnectedEndpointsType      string // Always "dcim.interface" or null in my sample.
	//  Duplex  unused in my sample
	//  L2vpnTermination  unused in my sample
	//  Mode  unused in my sample
	//  Module unused in my sample
	//  RfChannel
	//  RfChannelFrequency
	//  RfChannelWidth
	//  RfRole
	//  WWN
}

type Interfaces []*Interface

func (ints Interfaces) ForDeviceID(id int64) Interfaces {
	ret := []*Interface{}

	for _, i := range ints {
		if i.DeviceID == id {
			ret = append(ret, i)
		}
	}

	return ret
}

func dcimInterfaceToInterface(i *models.Interface) (*Interface, error) {
	in := &Interface{
		CableEnd:                    i.CableEnd,
		CustomFields:                make(map[string]interface{}),
		ConnectedEndpoints:          []string{},
		ConnectedEndpointsReachable: Bool(i.ConnectedEndpointsReachable),
		Description:                 i.Description,
		DeviceID:                    i.Device.ID,
		Display:                     i.Display,
		Enabled:                     i.Enabled,
		ID:                          i.ID,
		Label:                       i.Label,
		LinkPeers:                   []string{},
		LinkPeersType:               i.LinkPeersType,
		MACAddress:                  String(i.MacAddress),
		MarkConnected:               i.MarkConnected,
		MgmtOnly:                    i.MgmtOnly,
		MTU:                         Int64(i.Mtu),
		Name:                        String(i.Name),
		Speed:                       Int64(i.Speed),
		Tags:                        make(map[string]bool),
	}

	// if dev.CustomFields != nil {
	// }

	if i.Cable != nil {
		in.CableID = i.Cable.ID
		in.CableDisplay = i.Cable.Display
		in.CableLabel = i.Cable.Label
	}
	//	for _, ep := range i.ConnectedEndpoints {
	//		in.ConnectedEndpoints = append(in.ConnectedEndpoints, String(ep))
	//	}
	if i.Lag != nil {
		in.LAGName = String(i.Lag.Name)
	}

	//for _, lp := range i.LinkPeers {
	//  in.LinkPeers = append(in.LinkPeers, String(lp))
	//}

	if i.Parent != nil {
		in.ParentID = i.Parent.ID
		in.ParentName = String(i.Parent.Name)
	}

	for _, t := range i.Tags {
		in.Tags[*t.Name] = true
	}

	if i.Type != nil {
		in.Type = String(i.Type.Value)
	}

	if i.Vrf != nil {
		in.VRF = String(i.Vrf.Name)
	}

	return in, nil
}

func ListInterfaces(c *client.NetBoxAPI) (Interfaces, error) {
	var limit int64
	limit = 0

	r := dcim.NewDcimInterfacesListParams()
	r.Limit = &limit

	rs, err := c.Dcim.DcimInterfacesList(r, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to call DcimInterfacesList: %v", err)
	}

	ifs := make([]*Interface, len(rs.Payload.Results))
	for i, result := range rs.Payload.Results {
		in, err := dcimInterfaceToInterface(result)
		if err != nil {
			return nil, fmt.Errorf("Unable to convert dcim.Interface to netbox.Interface: %v", err)
		}

		ifs[i] = in
	}

	return ifs, nil
}

func GetInterface(c *client.NetBoxAPI, id int64) (*Interface, error) {
	var limit int64
	limit = 0
	idStr := string(id)

	r := dcim.NewDcimInterfacesListParams()
	r.Limit = &limit
	r.ID = &idStr

	rs, err := c.Dcim.DcimInterfacesList(r, nil)
	if err != nil {
		return nil, err
	}

	return dcimInterfaceToInterface(rs.Payload.Results[0])
}
