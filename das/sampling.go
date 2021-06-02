package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/lazyledger/lazyledger-core/ipfs"
	"github.com/lazyledger/lazyledger-core/libs/log"
	"github.com/lazyledger/lazyledger-core/p2p/ipld"
	ctypes "github.com/lazyledger/lazyledger-core/rpc/core/types"
	tmclient "github.com/lazyledger/lazyledger-core/rpc/jsonrpc/client"
	"github.com/lazyledger/lazyledger-core/types"
	"github.com/lazyledger/nmt/namespace"
	"github.com/spf13/cobra"
)

func sampleCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "sample",
		Aliases: []string{"sample", "s"},
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := args[0]

			iterations, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			times, err := Sample(cmd.Context(), host, int(iterations))
			if err != nil {
				return err
			}

			// easy improvement would be to save the times as a file instead of
			// pushing them to Stdout (eventually to a local log files)
			fmt.Println("data-start--------------")
			for i, t := range times {
				fmt.Println(i, t.Milliseconds())
			}
			fmt.Println("data-end----------------")
			return nil
		},
	}
}

func Sample(ctx context.Context, host string, iterations int) ([]time.Duration, error) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	apiProvider := ipfs.Embedded(false, ipfs.DefaultConfig(), logger)

	dag, cls, err := apiProvider()
	if err != nil {
		fmt.Println(err, 2)
		os.Exit(1)
	}
	defer cls.Close()

	client, err := tmclient.New(host)
	if err != nil {
		fmt.Println(err, 1)
		os.Exit(1)
	}

	dah, err := getLatestDAH(ctx, client)
	if err != nil {
		fmt.Println(err, 3)
		os.Exit(1)
	}

	sampleTimes := make([]time.Duration, iterations)
	for i := 0; i < iterations; i++ {
		start := time.Now()
		err = ipld.ValidateAvailability(ctx, dag, &dah, 1, func(pd namespace.PrefixedData8) {})
		if err != nil {
			fmt.Println(err, 4)
			os.Exit(1)
		}
		end := time.Now()
		total := end.Sub(start)
		fmt.Printf("#DATA sample %d %dms\n", i, total.Milliseconds())
		sampleTimes = append(sampleTimes, total)
	}

	return sampleTimes, nil
}

func getLatestDAH(ctx context.Context, client *tmclient.Client) (types.DataAvailabilityHeader, error) {
	height, err := getLatestHeight(ctx, client)
	if err != nil {
		return types.DataAvailabilityHeader{}, err
	}
	return getDAH(ctx, client, height)
}

func getDAH(ctx context.Context, client *tmclient.Client, height int64) (types.DataAvailabilityHeader, error) {
	var dah ctypes.ResultDataAvailabilityHeader
	_, err := client.Call(context.Background(), "data_availability_header", map[string]interface{}{"height": height}, &dah)
	if err != nil {
		return types.DataAvailabilityHeader{}, err
	}
	return dah.DataAvailabilityHeader, nil
}

func getLatestHeight(ctx context.Context, client *tmclient.Client) (int64, error) {
	var res ctypes.ResultBlockchainInfo
	_, err := client.Call(context.Background(), "blockchain", map[string]interface{}{}, &res)
	if err != nil {
		return 0, err
	}
	return res.LastHeight, nil
}
