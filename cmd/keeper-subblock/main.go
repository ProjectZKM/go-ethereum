package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	zkruntime "github.com/ProjectZKM/Ziren/crates/go-runtime/zkvm_runtime"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

// SubblockPayloadBytes matches the Rust host encoding model:
// it carries raw RLP bytes for the block and witness, plus `TxEnd`.
//
// The block and witness bytes are identical to `cmd/keeper`'s payload generation.
// This keeps the host-side encoding simple (Rust just concatenates bytes).
type SubblockPayloadBytes struct {
	ChainID    uint64
	BlockRLP   []byte
	WitnessRLP []byte
	TxEnd      uint64
}

type SubblockOutput struct {
	TxEnd           uint64
	GasUsed         uint64
	StateRoot       [32]byte
	ReceiptRoot     [32]byte
	ExecutedTxCount uint64
}

func init() {
	debug.SetGCPercent(-1)
}

func main() {
	payload := zkruntime.Read[SubblockPayloadBytes]()

	var block types.Block
	if err := rlp.DecodeBytes(payload.BlockRLP, &block); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode block rlp: %v\n", err)
		os.Exit(15)
	}
	var witness stateless.Witness
	if err := rlp.DecodeBytes(payload.WitnessRLP, &witness); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode witness rlp: %v\n", err)
		os.Exit(16)
	}

	txs := block.Transactions()
	if payload.TxEnd > uint64(len(txs)) {
		fmt.Fprintf(os.Stderr, "payload: tx_end out of range: %d > %d\n", payload.TxEnd, len(txs))
		os.Exit(17)
	}

	chainConfig, err := getChainConfig(payload.ChainID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get chain config: %v\n", err)
		os.Exit(13)
	}
	vmConfig := vm.Config{}

	// Build a prefix block [0:TxEnd).
	prefixTxs := make([]*types.Transaction, payload.TxEnd)
	copy(prefixTxs, txs[:payload.TxEnd])
	body := &types.Body{
		Transactions: prefixTxs,
		Uncles:       block.Uncles(),
		Withdrawals:  block.Withdrawals(),
	}
	header := block.Header()
	// Stateless mode ignores state/receipt roots, but still checks GasUsed and Bloom.
	// We'll validate against a rebuilt block after processing.
	header.Root = common.Hash{}
	header.ReceiptHash = common.Hash{}
	header.GasUsed = 0
	header.Bloom = types.Bloom{}

	prefixBlock := types.NewBlock(header, body, nil, trie.NewStackTrie(nil))
	isLast := payload.TxEnd == uint64(len(txs))

	stateRoot, receiptRoot, res, err := core.ExecuteStatelessWithResultSubblock(
		context.Background(),
		chainConfig,
		vmConfig,
		prefixBlock,
		&witness,
		isLast,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stateless execution failed: %v\n", err)
		os.Exit(10)
	}

	out := SubblockOutput{
		TxEnd:           payload.TxEnd,
		GasUsed:         res.GasUsed,
		StateRoot:       stateRoot,
		ReceiptRoot:     receiptRoot,
		ExecutedTxCount: uint64(len(res.Receipts)),
	}
	zkruntime.Commit(out)
	// Stock Go's normal exit does NOT route through zkvm.RuntimeExit, so the
	// public-values SHA-256 digest (committed_value_digest) would never be
	// committed. Explicitly run RuntimeExit to commit it (and re-commit the
	// deferred digest, which is idempotent) and HALT cleanly.
	zkruntime.RuntimeExit(0)
}
