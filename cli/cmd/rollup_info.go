package cmd

import (
	"encoding/json"
	"fmt"
	"hummingbird/config"
	"hummingbird/node"
	"hummingbird/node/contracts"
	"hummingbird/rollup"
	"hummingbird/utils"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func init() {
	RollupInfoCmd.Flags().Bool("json", false, "output info in json format")
	RollupInfoCmd.Flags().String("hash", "", "block hash to get info for")
	RollupInfoCmd.Flags().Uint64("num", 0, "block number to get info for")
}

var RollupInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "info will print information about the current rollup state",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		logger := ConsoleLogger()
		ethKey := getEthKey()
		useJson, _ := cmd.Flags().GetBool("json")

		n, err := node.NewFromConfig(cfg, logger, ethKey)
		utils.NoErr(err)

		r := rollup.NewRollup(n, &rollup.Opts{
			L1PollDelay:           time.Duration(cfg.Rollup.L1PollDelay) * time.Millisecond,
			L2PollDelay:           time.Duration(cfg.Rollup.L2PollDelay) * time.Millisecond,
			BundleSize:            cfg.Rollup.BundleSize,
			StoreCelestiaPointers: cfg.Rollup.StoreCelestiaPointers,
			StoreHeaders:          cfg.Rollup.StoreHeaders,
			Logger:                logger.With("ctx", "Rollup"),
		})

		var useHash bool

		// is a hash specified?
		hash, err := cmd.Flags().GetString("hash")
		useHash = err == nil && hash != ""
		// is a number specified?
		num, err := cmd.Flags().GetUint64("num")
		useNum := err == nil

		// if a hash is specified, get info for the block with that hash
		if useHash {
			info, err := r.GetBlockInfo(common.HexToHash(hash))
			utils.NoErr(err)
			printInfo(info, useJson)
			return
		}

		if useNum {
			h, err := r.Ethereum.GetRollupHeader(num)
			utils.NoErr(err)
			hash, err := contracts.HashCanonicalStateChainHeader(&h)
			utils.NoErr(err)
			info, err := r.GetBlockInfo(hash)
			utils.NoErr(err)
			printInfo(info, useJson)
			return
		}

		// otherwise get info for the chain
		info, err := r.GetInfo()
		utils.NoErr(err)
		printInfo(info, useJson)
	},
}

func printInfo(info any, useJson bool) {
	if useJson {
		buf, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(buf))
		return
	}

	// otherwise print as pretty text
	fmt.Println(" ")
	fmt.Println(utils.MarshalText(info))
	fmt.Println(" ")
}
