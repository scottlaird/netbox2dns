# netbox2dns

netbox2dns is a tool for publishing DNS records from [Netbox](http://netbox.dev) data.

Netbox provides a reasonable interface for managing and documenting IP
addresses and network devices, but out of the box there's no good way
to publish Netbox's data into DNS.  This tool is designed to publish
A, AAAA, and PTR records from Netbox into Google Cloud DNS.  It should
be possible to add other DNS providers without too much work, as long
as they're able to handle incremental record additions and removals.

## Compiling

Check out a copy of the `netbox2dns` code from GitHub using `git clone
https://github.com/scottlaird/netbox2dns.git`.  Then, run 'go build
cmd/netbox2dns/netbox2dns.go`, and it should generate a `netbox2dns`
binary.  This can be copied to other directories or other systems as
needed.

## Configuration

Edit `netbox2dns.yaml`.  Here is an example config:

```yaml
config:
  netbox: 
    host:  "netbox.example.com"
    token: "01234567890abcdef"

  defaults:
    project: "google-cloud-dns-project-name-123456"
    ttl: 300
  
  zones: 
    - name: "internal.example.com"
      zonename: "internal-example-com"
    - name: "example.com"
      zonename: "example-com"
    - name: "10.in-addr.arpa"
      zonename: "reverse-v4-10"
      delete_entries: true
    - name: "0.0.0.0.ip6.arpa"
      zonename: "reverse-v6-0000"
      delete_entries: true
```

To talk to Netbox, you'll need to provide your Netbox host, a Netbox
API token with (at a minimum) read access to Netbox's IP Address data.

To talk to Google Cloud DNS, you'll need to specify a project ID.
This should match the Google Cloud project name that hosts your DNS
records on console.cloud.google.com.  For now, netbox2dns uses
[Application Default
Credentials](https://cloud.google.com/docs/authentication/application-default-credentials).
See Google's documentation for how to set these up using the `gcloud`
CLI.

Finally, list your zones. When adding new records, netbox2dns will add
records to the *longest* matching zone name.  For the example above,
with `internal.example.com` and `example.com`, if Netbox has a record
for `router1.internal.example.com`, then it will be added to
`internal.example.com`.  Any records that don't fix into a listed zone
will be ignored.

By default, netbox2dns will search in `/etc/netbox2dns/`,
`/usr/local/etc/netbox2dns/`, and the correct directory for its config
file.  Config files can be in YAML (shown above), JSON, or CUE format.
Examples in [all 3
formats](https://github.com/scottlaird/netbox2dns/tree/main/testdata/config4)
are available.

## Use

Upon startup, netbox2dns will fetch all IP Address records from
Netbox *and* all A/AAAA/PTR records from the listed zones.

For each address in Netbox that is in state `active` and has a
non-empty DNS name, netbox2dns will attempt to add a forward DNS
record from the DNS name to the address, and a reverse DNS record from
the address to the DNS name.  IPv4 and IPv6 should be handled
automatically.

netbox2dns will then show a diff between the A/AAAA/PTR records found
in the existing zones and the records generated from Netbox.  If the
`--dry_run=false` flag is set, then it will add missing records, but
will not remove records from DNS that are not in Netbox.  If the
`delete_entries: true` setting is enabled for a zone, then netbox2dns
will remove any unknown A, AAAA, or PTR records from Google Cloud DNS.
This makes the most sense for reverse DNS, when Netbox is the source
of truth for all IP address assignement.
