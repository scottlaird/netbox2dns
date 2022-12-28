package netbox

import (
	"net/netip"

	"github.com/netbox-community/go-netbox/netbox/client"
	"github.com/netbox-community/go-netbox/netbox/client/ipam"
	"github.com/netbox-community/go-netbox/netbox/models"
)

// Extracted from netbox/models/device's Device, but reduced to basic
// Go types to make it easier to work with the subset that I care
// about.
type IPAddr struct {
	CustomFields       map[string]interface{}
	Address            netip.Prefix
	AssignedObjectID   int64
	AssignedObjectType string
	Description        string
	Display            string
	DNSName            string
	Family             string
	ID                 int64
	Role               string
	Status             string
	Tags               map[string]bool // Tags.Name -> true
	VRF                string
}

type IPAddrs []*IPAddr

func (ips IPAddrs) ForInterfaceID(id int64) IPAddrs {
	ret := IPAddrs{}

	for _, i := range ips {
		if i.AssignedObjectType == "dcim.interface" && i.AssignedObjectID == id {
			ret = append(ret, i)
		}
	}

	return ret
}

func ipamAddressToIPAddr(i *models.IPAddress) (*IPAddr, error) {
	ip := &IPAddr{
		CustomFields:       make(map[string]interface{}),
		AssignedObjectID:   Int64(i.AssignedObjectID),
		AssignedObjectType: String(i.AssignedObjectType),
		Description:        i.Description,
		Display:            i.Display,
		DNSName:            i.DNSName,
		ID:                 i.ID,
		Tags:               make(map[string]bool),
	}

	if i.Address != nil {
		prefix, err := netip.ParsePrefix(String(i.Address))
		if err != nil {
			return nil, err
		}
		ip.Address = prefix
	}
	if i.Family != nil {
		ip.Family = String(i.Family.Label)
	}
	if i.Role != nil {
		ip.Role = String(i.Role.Value)
	}
	if i.Status != nil {
		ip.Status = String(i.Status.Value)
	}
	if i.Vrf != nil {
		ip.VRF = String(i.Vrf.Name)
	}
	for _, t := range i.Tags {
		ip.Tags[*t.Name] = true
	}

	return ip, nil
}

func ListIPAddrs(c *client.NetBoxAPI) (IPAddrs, error) {
	var limit int64
	limit = 0

	r := ipam.NewIpamIPAddressesListParams()
	r.Limit = &limit

	rs, err := c.Ipam.IpamIPAddressesList(r, nil)
	if err != nil {
		return nil, err
	}

	ips := make(IPAddrs, len(rs.Payload.Results))
	for i, result := range rs.Payload.Results {
		ip, err := ipamAddressToIPAddr(result)
		if err != nil {
			return nil, err
		}

		ips[i] = ip
	}

	return ips, nil
}

func GetIPAddr(c *client.NetBoxAPI, id int64) (*IPAddr, error) {
	var limit int64
	limit = 0
	idStr := string(id)

	r := ipam.NewIpamIPAddressesListParams()
	r.Limit = &limit
	r.ID = &idStr

	rs, err := c.Ipam.IpamIPAddressesList(r, nil)
	if err != nil {
		return nil, err
	}

	return ipamAddressToIPAddr(rs.Payload.Results[0])
}
