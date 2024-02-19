package cmd

import (
	"fmt"
	"hummingbird/config"
	"hummingbird/defender"
	"hummingbird/node"
	"hummingbird/utils"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	DefenderProvideCmd.Flags().String("type", "header", "type of data to provide (header, tx)")
}

var DefenderProvideCmd = &cobra.Command{
	Use:   "provide",
	Short: "provide will download data from Celestia and provide it to Layer 1",
	Aliases: []string{
		"rblock",
		"hash",
	},
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		logger := GetLogger(viper.GetString("log-type"))
		ethKey := getEthKey()

		rblockHash := common.HexToHash(args[0])

		n, err := node.NewFromConfig(cfg, logger, ethKey)
		utils.NoErr(err)

		// allow block hash or number
		targetHash := common.HexToHash(args[1])
		var l2blockHash common.Hash
		if strings.HasPrefix(args[1], "0x") {
			l2blockHash = common.HexToHash(args[1])
		} else {
			logger.Info("Providing L2 Header by block number", "block", args[1])
			num, err := strconv.Atoi(args[1])
			utils.NoErr(err)
			b, err := n.LightLink.GetBlock(uint64(num))
			utils.NoErr(err)

			h := b.Header()
			h.Extra = common.Hex2Bytes("0x")
			l2blockHash = h.Hash()
			logger.Info("Providing L2 Header by block number", "block", args[1], "hash", l2blockHash.Hex())
		}

		d := defender.NewDefender(n, &defender.Opts{
			Logger: logger.With("ctx", "Defender"),
		})

		// type
		t, _ := cmd.Flags().GetString("type")
		var tx *types.Transaction
		switch t {
		case "header":
			logger.Info("Providing L2 Header...")
			tx, err = d.ProvideL2Header(rblockHash, targetHash)
			if err != nil {
				logger.Error("Defender.Provide failed", "err", err)
			}
		case "tx":
			logger.Info("Providing L2 Tx...")
			tx, err = d.ProvideL2Tx(rblockHash, targetHash)
			if err != nil {
				logger.Error("Defender.Provide failed", "err", err)
			}
		default:
			logger.Error("Invalid type", "type", t)
			return
		}

		fmt.Println(" ")
		fmt.Println("Tx Hash:", tx.Hash().Hex())
		fmt.Println("Provided L2 Data:", targetHash.Hex())
		fmt.Println("Included in Rollup Block:", rblockHash.Hex())
		fmt.Println(" ")
	},
}
