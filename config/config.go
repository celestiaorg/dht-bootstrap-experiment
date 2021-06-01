package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/digitalocean/godo"
	do "github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
)

// Config structures the data used to configure a deployment
type Config struct {
	Droplets map[string]Droplet `json:"droplets"`
	// SSHKeyID is the fingerprint of the ssh key preloaded into digital ocean
	SSHKeyID string `json:"ssh_key_id"`
	// Tag is used to idendify droplets that belong to this deployment
	Tag string `json:"tag"`
}

// Droplet specifies each droplet
type Droplet struct {
	// Location is the digital ocean region for the droplet
	// ie "nyc3"
	Location string `json:"location"`
	// Size indicates how big the droplet size should be
	// ie "s-1vcpu-1gb"
	Size string `json:"size"`
	// Type: Validator || Full || LightClient || DHT
	Type NodeType `json:"droplet_type"`
	// Payload is the path to directory that is to be copied to the server
	Payload string `json:"payload"`
	// InitCommands are the ssh commands run after payload is delivered
	InitCommands []string `json:"init_commands"`
	// Peers contains the names of the other droplets to connect to
	Peers []string
	// Output is the path to output file
	Output string `json:"output"`
	Drop   godo.Droplet
}

type NodeType int

const (
	Validator NodeType = iota
	Full
	LightClient
	DHT
)

func (n NodeType) String() string {
	switch n {
	case Validator:
		return "Validator"
	case Full:
		return "Full"
	case LightClient:
		return "LightClient"
	case DHT:
		return "DHT"
	default:
		// panic here?
		return "unrecognized node type"
	}
}

// Match connects each digital ocean droplet with the configured Drop
func (c Config) Match(ctx context.Context, client *godo.Client) (Config, error) {
	drops, err := DropletList(ctx, client)
	if err != nil {
		return c, err
	}

	drops, err = FilterDrops(drops, c.Tag)
	if err != nil {
		return c, err
	}

	for _, drop := range drops {
		confDrop, has := c.Droplets[drop.Name]
		if !has {
			return c, fmt.Errorf("droplet not found in config: %s", drop.Name)
		}
		confDrop.Drop = drop
		c.Droplets[drop.Name] = confDrop
	}

	return c, nil
}

func defaultConfig() Config {
	return Config{
		Droplets: map[string]Droplet{
			"validator1": {
				Location: string(do.RegionNYC3),
				Size:     string(do.DropletSlugDropletS1VCPU1GB),
				Type:     0,
			},
		},
		Tag:      "devnet",
		SSHKeyID: "put-do-ssh-key-finger-print-here",
	}
}

func LoadConfig(path string) (Config, error) {
	configData, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var conf Config
	err = json.Unmarshal(configData, &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, conf.ValidateBasic()
}

func (c Config) ValidateBasic() error {
	if len(c.Droplets) == 0 {
		return errors.New("no droplets configured")
	}
	if len(c.SSHKeyID) == 0 {
		return errors.New("no ssh key finger print provided")
	}
	for name, drop := range c.Droplets {
		for _, peer := range drop.Peers {
			if _, has := c.Droplets[peer]; !has {
				return fmt.Errorf(
					"%s has a peer, %s, who is not defined in the Config", name, peer,
				)
			}
		}
	}
	return nil
}

func WriteConfig(path string, config Config) error {
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, configData, 0700)
}

// WriteIPsJson collects all of the public IPv4s of the existing droplets and writes
// them to a json file in each unique payload path
func (c Config) WriteIPsJson() error {
	ips, err := c.exportIPs()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(ips, "", "  ")
	if err != nil {
		return err
	}
	uniquePayloadPaths := make(map[string]struct{})
	for _, drop := range c.Droplets {
		uniquePayloadPaths[drop.Payload] = struct{}{}
	}
	for payPath := range uniquePayloadPaths {
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s", payPath, "public_ipv4s.json"), data, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c Config) WriteIPsBash() error {
	const bashTemplate = `
#!/bin/bash
{{ range $name, $ip := . }}
export {{$name}}="{{$ip}}"
{{ end }}
`
	t, err := template.New("baships").Parse(bashTemplate)
	if err != nil {
		return err
	}
	ips, err := c.exportIPs()
	if err != nil {
		return err
	}
	uniquePayloadPaths := make(map[string]struct{})
	for _, drop := range c.Droplets {
		uniquePayloadPaths[drop.Payload] = struct{}{}
	}
	for payPath := range uniquePayloadPaths {
		file, err := os.OpenFile(fmt.Sprintf("%s/%s", payPath, "public_ipv4s.sh"), os.O_CREATE|os.O_WRONLY, 0700)
		if err != nil {
			return err
		}
		defer file.Close()
		err = t.Execute(file, ips)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c Config) exportIPs() (map[string]string, error) {
	out := make(map[string]string)
	for name, drop := range c.Droplets {
		ipv4, err := drop.Drop.PublicIPv4()
		if err != nil {
			return nil, err
		}
		out[name] = ipv4
	}
	return out, nil
}
