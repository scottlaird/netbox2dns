config: {
	netbox: {
		host:  "netbox.example.com"
		token: "changeme"
	}

	defaults: {
		project: "random-string"
		ttl:     300
	}

	zones: [
		{name: "internal.example.com"
			zonename: "internal-example-com"
 			zonetype: "clouddns"
		},
		{name: "example.com"
			zonename: "example-com"
 			zonetype: "clouddns"
		},
		{name: "10.in-addr.arpa"
			zonename:       "reverse-v4-10"
			delete_entries: true
 			zonetype: "clouddns"
		},
		{name: "0.0.0.0.ip6.arpa"
			zonename:       "reverse-v6-0000"
			delete_entries: true
 			zonetype: "clouddns"
		},
	]
}
