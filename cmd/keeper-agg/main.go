package main

import (
	"fmt"
	"os"
	"runtime/debug"

	zkruntime "github.com/ProjectZKM/Ziren/crates/go-runtime/zkvm_runtime"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type SubblockOutput struct {
	TxEnd           uint64
	GasUsed         uint64
	StateRoot       common.Hash
	ReceiptRoot     common.Hash
	ExecutedTxCount uint64
}

type AggregationPayloadBytes struct {
	ChainID    uint64
	BlockRLP   []byte
	Subblocks  []SubblockOutput
}

type AggregationOutput struct {
	Ok            bool
	FinalState    common.Hash
	FinalReceipts common.Hash
	SubblockCount uint64
}

func init() {
	debug.SetGCPercent(-1)
}

func main() {
	payload := zkruntime.Read[AggregationPayloadBytes]()

	var block types.Block
	if err := rlp.DecodeBytes(payload.BlockRLP, &block); err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode block rlp: %v\n", err)
		os.Exit(15)
	}
	if len(payload.Subblocks) == 0 {
		fmt.Fprintf(os.Stderr, "payload: no subblocks\n")
		os.Exit(16)
	}

	totalTxs := uint64(len(block.Transactions()))
	var prevTxEnd uint64
	for i, sb := range payload.Subblocks {
		if sb.TxEnd < prevTxEnd {
			fmt.Fprintf(os.Stderr, "subblocks: non-monotonic tx_end at %d\n", i)
			os.Exit(20)
		}
		if sb.TxEnd > totalTxs {
			fmt.Fprintf(os.Stderr, "subblocks: tx_end out of range at %d\n", i)
			os.Exit(21)
		}
		prevTxEnd = sb.TxEnd
	}

	last := payload.Subblocks[len(payload.Subblocks)-1]
	if last.TxEnd != totalTxs {
		fmt.Fprintf(os.Stderr, "final subblock does not reach end: tx_end=%d total=%d\n", last.TxEnd, totalTxs)
		os.Exit(22)
	}
	// The aggregation check for correctness: final roots must match the block header.
	if last.StateRoot != block.Root() {
		fmt.Fprintf(os.Stderr, "final state root mismatch: got=%x want=%x\n", last.StateRoot, block.Root())
		os.Exit(23)
	}
	if last.ReceiptRoot != block.ReceiptHash() {
		fmt.Fprintf(os.Stderr, "final receipt root mismatch: got=%x want=%x\n", last.ReceiptRoot, block.ReceiptHash())
		os.Exit(24)
	}

	out := AggregationOutput{
		Ok:            true,
		FinalState:    last.StateRoot,
		FinalReceipts: last.ReceiptRoot,
		SubblockCount: uint64(len(payload.Subblocks)),
	}
	zkruntime.Commit(out)
}
