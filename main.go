package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/digitalocean/godo"
	"github.com/evan-forbes/devnet/config"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cobra.Command{
		Use:     "root command",
		Aliases: []string{"devnet"},
	}

	rootCmd.AddCommand(InitCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "initialize deployment",
		Aliases: []string{"init", "i"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// get the digital ocean token from the env vars
			doat := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")

			// create the digital ocean client
			client := godo.NewFromToken(doat)

			// load the config from the working dir
			conf, err := config.LoadConfig("config.json")
			if err != nil {
				return err
			}

			// connect each existing do droplet to the configered ones
			conf, err = conf.Match(context.TODO(), client)
			if err != nil {
				return err
			}

			// establish ssh connections to each droplet
			manager, err := NewSSHManager(conf.Droplets)
			if err != nil {
				return err
			}
			defer manager.CloseAll()

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
						fmt.Println("ran comnmand", command, "on", n)
					}

				}(name, conn)
			}

			wg.Wait()
			return nil
		},
	}
}
