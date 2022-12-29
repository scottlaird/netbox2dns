// This defines the configuration format for netbox2dns, along with a
// validation rules for each field.  See http://cuelang.org for
// documenation.

// A #CloudDNSZone is a DNS zone hosted on Google Cloud DNS.
// Each field has a type ("string"), optionally a default (*),
// and some constraints.
#CloudDNSZone: {
	zonetype:        "clouddns"
	name:            string
	zonename:        string
	project:         *config.defaults.project | string
	ttl:             *config.defaults.ttl | int & >60 & <=86400
	delete_entries?: *false | bool // Remove entries that are missing
	...
}

#ZoneFileZone: {
	zonetype:        "zonefile"
	name:            string
	filename:        string
	ttl:             *config.defaults.ttl | int & >60 & <=86400
	delete_entries?: *false | bool // Remove entries that are missing
	...
}

#Zone: #CloudDNSZone | #ZoneFileZone

// This is the template for the actual configuration.
config: {
	// At least one zone is required.
	zones: [#Zone, ...#Zone]

	// Zonemap is generated internally and doesn't appear in
	// the YAML config file, etc.  It contains the same data
	// as zones:, but it's a map of name -> zone data, which
	// is less convienent in the config file but more convienent
	// to use.
	zonemap: [string]: #Zone
	zonemap: {
		for z in zones {
			"\(z.name)": z
		}
	}

	// Netbox config settings.
	netbox: {
		host:  string
		token: string
	}

	// Defaults.  Notice the `*config.defaults.` clauses above, in #CloudDNSZone.
	defaults: {
		ttl:       *300 | int
		project?:  *"foo" | string
	}
}
