package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/digitalocean/godo"
	"github.com/evan-forbes/devnet/config"
	llconfig "github.com/lazyledger/lazyledger-core/config"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cobra.Command{
		Use:     "devnet",
		Aliases: []string{"devnet"},
	}

	rootCmd.AddCommand(
		InitCmd(),
		NewTendermintConfigCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "init",
		Aliases: []string{"init", "i"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get the digital ocean token from the env vars
			doat := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")

			// create the digital ocean client
			client := godo.NewFromToken(doat)

			// load the config from the working dir
			conf, err := config.LoadConfig(args[0])
			if err != nil {
				return err
			}

			// connect each existing do droplet to the configered ones
			conf, err = conf.Match(context.TODO(), client)
			if err != nil {
				return err
			}

			// fetch the ssh password if any
			sshPass := os.Getenv("SSH_PASS")
			switch sshPass {
			case "nil":
				sshPass = ""
			case "":
				fmt.Println(
					"password to ssh key (press enter for no password or alternatively export as SSH_PASS). export as 'nil' to ignore future requests",
				)
				fmt.Scanf(
					"%s",
					&sshPass,
				)
			default:

			}
			fmt.Println("setting ssh pass")

			// establish ssh connections to each droplet
			manager, err := NewSSHManager(conf.Droplets, sshPass)
			if err != nil {
				return err
			}
			fmt.Println("created new ssh manager")

			defer manager.CloseAll()

			err = conf.WriteIPs()
			if err != nil {
				return err
			}
			fmt.Println("wrote public IPs to payloads")

			// deliver the payloads via scp
			for name, conn := range manager.Conns {
				err = conn.DeliverPayload()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("delivered payload for: name", name)
			}

			// run initial commands and forward their Stdouts and Stderrs to local files
			var wg sync.WaitGroup
			for name, conn := range manager.Conns {
				wg.Add(1)
				go func(n string, c Connection) {
					defer wg.Done()
					for _, command := range c.drop.InitCommands {
						err = c.Run(command)
						if err != nil {
							log.Println(fmt.Errorf("failure to run command %s on %s: %w", command, n, err))
							continue
						}
						fmt.Println("ran command", command, "on", n)
					}

				}(name, conn)
			}

			wg.Wait()
			return nil
		},
	}
}

func NewTendermintConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "creates a default tendermint config at the path provided",
		Aliases: []string{"newLazyConfig", "i"},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) == 1 {
				path = args[0]
			}
			WriteTendermintConfig(path, llconfig.DefaultConfig())
			return nil
		},
	}
}

func LoadTendermintConfig(path string) (llconfig.Config, error) {
	var cfg llconfig.Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to load config from %q: %w", path, err)
	}
	return cfg, cfg.ValidateBasic()
}

func WriteTendermintConfig(path string, cfg *llconfig.Config) error {
	err := cfg.ValidateBasic()
	if err != nil {
		return err
	}
	llconfig.WriteConfigFile(path, cfg)
	return nil
}
