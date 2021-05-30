package main

import (
	"log"

	"github.com/evan-forbes/devnet/config"
	do "github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// read the ssh pub key and provide it to each droplet
// load the bash files
// run the files

func main() {
	// load config
	conf, err := config.LoadConfig("../config.json")
	if err != nil {
		log.Fatal("config file 'config.json' not found in working dir")
	}

	pulumi.Run(func(ctx *pulumi.Context) error {
		// create each droplet described in the config.
		for name, dropConf := range conf.Droplets {
			drop, err := do.NewDroplet(
				ctx,
				name,
				&do.DropletArgs{
					Image:   pulumi.String("ubuntu-20-04-x64"),
					Region:  pulumi.String(dropConf.Location),
					Size:    pulumi.String(dropConf.Size),
					SshKeys: pulumi.ToStringArray([]string{conf.SSHKeyID}),
					Name:    pulumi.String(name),
					// tag with the global conf.Tag, along with the type of node
					Tags: pulumi.ToStringArray([]string{conf.Tag, dropConf.Type.String()}),
				},
			)

			if err != nil {
				return err
			}

			ctx.Export(name, drop.Ipv4Address)
		}

		return nil
	})
}
