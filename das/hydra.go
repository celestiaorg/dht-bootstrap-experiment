package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	config "github.com/ipfs/go-ipfs-config"
	fsrepo "github.com/ipfs/go-ipfs-config/serialize"
	"github.com/lazyledger/lazyledger-core/ipfs"
	"github.com/lazyledger/lazyledger-core/libs/log"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "init",
		Aliases: []string{"init"},
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

			apiProvider := ipfs.Embedded(true, ipfs.DefaultConfig(), logger)

			_, cls, err := apiProvider()
			if err != nil {
				fmt.Println(err, 2)
				os.Exit(1)
			}
			cls.Close()
			return nil
		},
	}
}

func addHydraCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "add-hydra [hydra-IP] [ipfs-config-path]",
		Aliases: []string{"add-hydra"},
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			hydraIP, cfgPath := args[0], args[1]
			return AddHydraIDs(cmd.Context(), hydraIP, cfgPath)
		},
	}
}

func AddHydraIDs(ctx context.Context, hydraIP, cfgPath string) error {
	rawIDs, err := getSwarmPeerIDs(context.Background(), fmt.Sprintf("http://%s/heads", hydraIP))
	if err != nil {
		return err
	}

	ids := rawIDs.IDs()

	var cfg config.Config
	err = fsrepo.ReadConfigFile(cfgPath, &cfg)
	if err != nil {
		return err
	}
	boots := make([]string, 0)
	boots = append(cfg.Bootstrap, ids...)
	cfg.Bootstrap = boots

	return fsrepo.WriteConfigFile(cfgPath, cfg)
}

type hydraHeadsResp struct {
	Addrs []string `json:"Addrs"`
	ID    string   `json:"ID"`
}

func (h hydraHeadsResp) IDs() []string {
	out := make([]string, len(h.Addrs))
	for i, addr := range h.Addrs {
		fmt.Println(fmt.Sprintf("%s/p2p/$%s", addr, h.ID))
		out[i] = fmt.Sprintf("%s/p2p/$%s", addr, h.ID)
	}
	return out
}

func getSwarmPeerIDs(ctx context.Context, url string) (hydraHeadsResp, error) {
	var ids hydraHeadsResp
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ids, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ids, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ids, err
	}

	err = json.Unmarshal(data, &ids)
	if err != nil {
		return ids, err
	}

	return ids, nil
}
