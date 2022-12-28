// This defines the configuration format for netbox2dns, along with a
// validation rules for each field.  See http://cuelang.org for
// documenation.

// A #CloudDNSZone is a DNS zone hosted on Google Cloud DNS.
#CloudDNSZone: {
	zonetype:        *"clouddns" | string
	name:            string
	zonename:        string
	project:         *config.defaults.project | string
	ttl:             *config.defaults.ttl | int & >60 & <=86400
	delete_entries?: *false | bool // Remove entries that are missing
}

// To add other DNS providers, you'll need to add a similar
// definition.  For example, #Route53Zone for AWS.  The zonetype, name,
// ttl, and delete_entries fields are required.  Add any other fields
// needed to define zones for your provider, similar to `project` and
// `zonename`.  You will also need to update config.go with any new fields,
// and then add code for talking to your provider in dns.go and elsewhere.
// You'll probably want to use clouddns.go as an example.

// #Zone should be a union of all supported Zone types.  If you add
// new providers, then you'll probably need to add something like `|
// Route53Zone` at the end of the line.

#Zone: #CloudDNSZone

// This is the template for the actual configuration.
#Config: {
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
		zonetype?: *"clouddns" | string
		ttl:       *300 | int
		project?:  *"foo" | string
	}
}

config: #Config
