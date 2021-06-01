package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	llconfig "github.com/lazyledger/lazyledger-core/config"
	"github.com/spf13/cobra"
)

const defaultConfigPath = "/root/.tendermint/config/config.toml"

func main() {
	rootCmd := cobra.Command{
		Use:     "setup",
		Aliases: []string{"setup"},
	}

	rootCmd.AddCommand(
		addPeerCmd(),
		openListeningRPC(),
		newTendermintConfigCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addPeerCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "add a persistent peer to the default tendermint config",
		Aliases: []string{"addpeer", "add"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			cfg, err := loadLazyConfig(defaultConfigPath)
			if err != nil {
				return err
			}
			ips, err := readPubIPs("/root/validator/public_ipv4s.json")
			if err != nil {
				return err
			}

			ip, has := ips[args[0]]
			if !has {
				return fmt.Errorf("no public IP for %s found", args[0])
			}

			addPersistentPeer(ip, 26656, cfg)
			cfg.LogFormat = "plain"
			cfg.TxIndex = &llconfig.TxIndexConfig{}
			saveLazyConfig(defaultConfigPath, cfg)
			return nil
		},
	}
}

func newTendermintConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "defconf",
		Aliases: []string{"defconf"},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "/root/.tendermint/config/config.toml"
			saveLazyConfig(path, llconfig.DefaultConfig())
			return nil
		},
	}
}

func openListeningRPC() *cobra.Command {
	return &cobra.Command{
		Use:     "open the RPC to the internetye",
		Aliases: []string{"openRPC"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadLazyConfig(defaultConfigPath)
			if err != nil {
				return err
			}

			cfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"
			cfg.LogFormat = "plain"
			cfg.TxIndex = &llconfig.TxIndexConfig{}

			saveLazyConfig(defaultConfigPath, cfg)
			return nil
		},
	}
}

func addPersistentPeer(peer string, port int, cfg *llconfig.Config) {
	peer = fmt.Sprintf("%s:%d", peer, port)
	switch cfg.P2P.PersistentPeers {
	case "":
		cfg.P2P.PersistentPeers = peer
	default:
		cfg.P2P.PersistentPeers = fmt.Sprintf("%s,%s", cfg.P2P.PersistentPeers, peer)
	}
}

func loadLazyConfig(path string) (*llconfig.Config, error) {
	var cfg llconfig.Config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %q: %w", path, err)
	}
	return &cfg, nil
}

func saveLazyConfig(path string, cfg *llconfig.Config) {
	llconfig.WriteConfigFile(path, cfg)
}

func readPubIPs(path string) (map[string]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ips := make(map[string]string)
	err = json.Unmarshal(data, &ips)
	if err != nil {
		return nil, err
	}
	return ips, nil
}
