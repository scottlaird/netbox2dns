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
		},
		{name: "example.com"
			zonename: "example-com"
		},
		{name: "10.in-addr.arpa"
			zonename:       "reverse-v4-10"
			delete_entries: true
		},
		{name: "0.0.0.0.ip6.arpa"
			zonename:       "reverse-v6-0000"
			delete_entries: true
		},
	]
}
