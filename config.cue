import "strings"

#CloudDNSZone: {
	zonetype:        *"clouddns" | string
        name:            string
	zonename:        string
	project:         *config.defaults.project | string
	ttl:             *config.defaults.ttl | int & >60 & <=86400
	delete_entries?: *false | bool // Remove entries that are missing
}

#Zone: #CloudDNSZone

#Config: {
        // Require at least one zone
	zones: [#Zone, ...#Zone]

	zonemap: [string]: #Zone
	zonemap: {
	  for z in zones {
		"\(z.name)": z
	  }
	  }
	
	netbox: {
		host: string
		token: string
	}

        defaults: {
		zonetype?: *"clouddns" | string
		ttl:       *300 | int
		project?:  *"foo" | string
	}
}

config: #Config