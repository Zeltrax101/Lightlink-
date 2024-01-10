package defender

import (
	"fmt"
	"hummingbird/node"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Opts struct {
	Logger *slog.Logger
	DryRun bool // DryRun indicates whether or not to actually submit the block to the L1 rollup contract.
}

type Defender struct {
	*node.Node
	Opts *Opts
}

func NewDefender(node *node.Node, opts *Opts) *Defender {
	return &Defender{Node: node, Opts: opts}
}

func (d *Defender) ProveDA(txHash common.Hash) (*node.CelestiaProof, error) {
	return d.Celestia.GetProof(txHash[:])
}

func (d *Defender) DefenderDA(block common.Hash, txHash common.Hash) (*types.Transaction, error) {
	proof, err := d.ProveDA(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to prove data availability: %w", err)
	}

	return d.Ethereum.DefendDataRootInclusion(block, proof)
}
