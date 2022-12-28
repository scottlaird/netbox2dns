package netbox2dns

import (
	"net/netip"
	"testing"
)

func TestAddZonesSorted(t *testing.T) {
	z := NewZones()

	if z.Zones == nil {
		t.Errorf("z.Zones is nil, want map[string]*Zone")
	}

	a := &Zone{
		Name: "a.com",
	}
	b := &Zone{
		Name: "bb.com",
	}
	c := &Zone{
		Name: "ccc.com",
	}

	z.AddZone(b)
	z.AddZone(a)
	z.AddZone(c)

	if len(z.Zones) != 3 {
		t.Fatalf("len(z.Zones) got %d, want 3", len(z.Zones))
	}
	if z.Zones["a.com"] == nil {
		t.Error("z.Zones[\"a.com\"] got nil, want !nil")
	}
	if z.Zones["bb.com"] == nil {
		t.Error("z.Zones[\"bb.com\"] got nil, want !nil")
	}
	if z.Zones["ccc.com"] == nil {
		t.Error("z.Zones[\"ccc.com\"] got nil, want !nil")
	}

	if len(z.sortedZones) != 3 {
		t.Errorf("len(z.sortedZones) got %d, want 3", len(z.sortedZones))
	}
	if z.sortedZones[0].Name != "ccc.com" {
		t.Errorf("z.sortedZones[0].Name: got %q want %q", z.sortedZones[0].Name, "ccc.com")
	}
	if z.sortedZones[1].Name != "bb.com" {
		t.Errorf("z.sortedZones[1].Name: got %q want %q", z.sortedZones[1].Name, "bb.com")
	}
	if z.sortedZones[2].Name != "a.com" {
		t.Errorf("z.sortedZones[2].Name: got %q want %q", z.sortedZones[2].Name, "a.com")
	}
}

func TestReverseName4(t *testing.T) {
	addr := netip.MustParseAddr("1.2.3.4")
	want := "4.3.2.1.in-addr.arpa."

	got := reverseName4(addr)
	if got != want {
		t.Errorf("reverseName4(%s) wrong, got %q want %q", addr.String(), got, want)
	}

	got = ReverseName(addr)
	if got != want {
		t.Errorf("ReverseName(%s) wrong, got %q want %q", addr.String(), got, want)
	}
}

func TestReverseName6(t *testing.T) {
	addr := netip.MustParseAddr("1:2::3:4")
	want := "4.0.0.0.3.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.2.0.0.0.1.0.0.0.ip6.arpa."

	got := reverseName6(addr)
	if got != want {
		t.Errorf("reverseName6(%s) wrong, got %q want %q", addr.String(), got, want)
	}

	got = ReverseName(addr)
	if got != want {
		t.Errorf("ReverseName(%s) wrong, got %q want %q", addr.String(), got, want)
	}
}
