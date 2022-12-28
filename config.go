package netbox2dns

import (
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/gocode/gocodec"
	"cuelang.org/go/encoding/yaml"
	"cuelang.org/go/encoding/json"

	log "github.com/golang/glog"

	_ "embed"
)

// type ConfigRoot matches the root of the schema defined in `config.cue`.
type ConfigRoot struct {
	Config Config `json:"config,omitempty"`
}

// type Config matches the `config` item in the schema defined in
// `config.cue`.  Each item must be marked as `notempty` and must have
// a JSON tag that matches the name in the CUE file.
type Config struct {
	Netbox struct {
		Host  string `json:"host,omitempty"`
		Token string `json:"token,omitempty"`
	} `json:"netbox,omitempty"`
	Defaults struct {
		Zonetype string `json:"zonetype,omitempty"`
		Ttl      int64  `json:"ttl,omitempty"`
		Project  string `json:"project,omitempty"`
	} `json:"defaults,omitempty"`
	ZoneMap map[string]*ConfigZone `json:"zonemap,omitempty"`
	Zones   []*ConfigZone          `json:"zones,omitempty"`
}

// type ConfigZone matches `Zone` in `config.cue`.  This needs to be
// the union of all defined zone types.  At the moment, this is only
// CloudDNSZone, but other types are possible.  They're switched based
// on the `ZoneType` field.  Then, code in `dns.go` uses that to
// dispatch to the correct back-end handler.
type ConfigZone struct {
	ZoneType      string `json:"zonetype,omitempty"`
	Name          string `json:"name,omitempty"`
	ZoneName      string `json:"zonename,omitempty"`
	Project       string `json:"project,omitempty"`
	Ttl           int64  `json:"ttl,omitempty"`
	DeleteEntries bool   `json:"delete_entries,omitempty"`
}

// This causes "config.cue" in the current directory to be embedded
// into the compiled Go code as "cueSchema".
//
//go:embed config.cue
var cueSchema []byte

// List of directories to search for config files.
var configDirs = []string{
	"/usr/local/etc/netbox2dns",
	"/etc/netbox2dns",
	".",
}

// List of supported config file extensions.
var configExtensions = []string{
	"yaml",
	"yml",
	"json",
	"cue",
}

// function FindConfig looks in several locations for a config file
// named "$basename.yml", "$basename.yaml", "$basename.json", or
// "$basename.cue".
func FindConfig(basename string) (string, error) {
	return findConfig(basename, configDirs, configExtensions)
}

func findConfig(basename string, configDirs []string, configExtensions []string) (string, error) {
	for _, d := range configDirs {
		for _, e := range configExtensions {
			p := fmt.Sprintf("%s/%s.%s", d, basename, e)
			log.Infof("Checking %q", p)
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("Config file %q not found in %+v with extension %+v", basename, configDirs, configExtensions)
}

// function ParseConfig parses a config file and returns a Config
// object or an error.  When used with FindConfig, it can hunt down a
// config file in several formats and then parse and validate it
// automatically.
func ParseConfig(filename string) (*Config, error) {
	config := &ConfigRoot{}

	cctx := cuecontext.New()

	if strings.HasSuffix(filename, ".yml") || strings.HasSuffix(filename, ".yaml") {
		err := parseYAML(filename, config, cctx)
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(filename, ".json") {
		err := parseJSON(filename, config, cctx)
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(filename, ".cue") {
		err := parseCUE(filename, config, cctx)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Unknown config format for %q", filename)
	}

	// Compile "config.cue"
	schema := cctx.CompileBytes(cueSchema)

	// Apply the schema to the results of the parsed YAML file.
	// This is basically equivalent to 'cue eval config.cue
	// config.yaml'.
	codec := gocodec.New((*cue.Runtime)(cctx), nil)
	err := codec.Complete(schema, config)
	if err != nil {
		return nil, err
	}

	return &(config.Config), nil
}

// parseYAML parses a YAML (.yml, .yaml) file into a ConfigRoot.
func parseYAML(filename string, cfg *ConfigRoot, cctx *cue.Context) error {
	// yaml.Extract will do the read itself if the second parameter is nil.
	yamlAST, err := yaml.Extract(filename, nil)
	if err != nil {
		return err
	}
	yamlValue := cctx.BuildFile(yamlAST)
	yamlValue.Decode(cfg)

	return nil
}

// parseJSON parses a JSON file into a ConfigRoot.
func parseJSON(filename string, cfg *ConfigRoot, cctx *cue.Context) error {
	// json.Extract will *not* do the read itself if the second
	// parameter is nil, unlike yaml.Extract.
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	jsonAST, err := json.Extract(filename, b)
	if err != nil {
		return err
	}
	jsonValue := cctx.BuildExpr(jsonAST)
	jsonValue.Decode(cfg)

	return nil
}

// parseCUE parses a .cue-format config file into a ConfigRoot
func parseCUE(filename string, cfg *ConfigRoot, cctx *cue.Context) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	config := cctx.CompileBytes(b)
	config.Decode(cfg)
	
	return nil
}
