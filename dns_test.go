package netbox2dns

import (
	"testing"
)

type serials struct {
	initial, want uint32
}

func TestIncrementSerial(t *testing.T) {
	tests := []serials{
		{1, 2},
		{2, 3},
		// {2000010099, 2000010100},  // Expected, but implementation-defined
		// {2000010100, 2022123000},  // Expected, but implementation-defined
		{2022122904, 2022123000},
		{2022123005, 2022123006},
		{2022123099, 2022123100},
		{2022123100, 0}, // Error, date is in the future
	}

	cz := &ConfigZone{}
	today := "20221230"

	for _, test := range tests {
		got, err := incrementSerialFixedDate(cz, test.initial, today)

		if test.want == 0 {
			if err == nil {
				t.Errorf("IncrementSerial(%d) should have returned an error but did not", test.initial)
			}
		} else {
			if err != nil {
				t.Errorf("IncrementSerial(%d) returned error: %v", test.initial, err)
			}
			if got != test.want {
				t.Errorf("IncrementSerial(%d): got %d want %d", test.initial, got, test.want)
			}
		}
	}
}
