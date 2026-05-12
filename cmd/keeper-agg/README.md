# keeper-agg (zkVM guest)

`keeper-agg` is a go-ethereum zkVM guest program that **aggregates and constrains** a sequence of
subblock results produced by `keeper-subblock`.

It is designed to be driven by a host-side pipeline that:
1) Computes subblock boundaries (matching the Rust pipeline semantics),
2) Executes/proves each `keeper-subblock` segment,
3) Feeds the collected subblock outputs (and optionally subblock proofs) into `keeper-agg`.

## Goal

- Check basic subblock sequencing constraints (monotonic `tx_end`, coverage of the full tx list).
- Constrain the *final* execution result by checking that the last subblock roots match the block
  header roots (`stateRoot` and `receiptsRoot`).

## Input (stdin)

With the `ziren` build tag enabled, the program reads a single struct from stdin via
`zkvm_runtime.Read[T]` (Ziren go-runtime "bincode-like" encoding compatible with Rust `ZKMStdin::write`):

```go
type SubblockOutput struct {
    TxEnd           uint64
    GasUsed         uint64
    StateRoot       common.Hash
    ReceiptRoot     common.Hash
    ExecutedTxCount uint64
}

type AggregationPayloadBytes struct {
    ChainID   uint64
    BlockRLP  []byte        // raw block RLP bytes (from debug_getRawBlock)
    Subblocks []SubblockOutput
}
```

Notes:
- `BlockRLP` must decode to `types.Block` and provide the canonical `Root()` and `ReceiptHash()`.
- `Subblocks` must be in the same order as the host segmentation.

## Output (public values)

The program commits one struct via `zkvm_runtime.Commit(AggregationOutput)`:

```go
type AggregationOutput struct {
    Ok            bool
    FinalState    common.Hash
    FinalReceipts common.Hash
    SubblockCount uint64
}
```

## What is (not yet) verified

Current implementation verifies:
- `tx_end` is non-decreasing and within `[0, total_txs]`
- The last subblock reaches the end (`tx_end == total_txs`)
- The last subblock `(stateRoot, receiptRoot)` matches the block header `(Root, ReceiptHash)`

Not verified yet (planned for proof-carrying aggregation):
- That each `SubblockOutput` actually comes from a valid `keeper-subblock` execution/proof.
  This requires zkVM-side deferred proof verification (feeding subblock proofs via the host into the
  proof stream and invoking the verify syscall from within `keeper-agg`).
- Linking constraints between consecutive subblocks (e.g. chaining intermediate roots), if/when
  those semantics are required by the pipeline.

## Build

```bash
cd cmd/keeper-agg
GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -tags "ziren" ./...
```

Environment (typically provided by the host build script / overlay):
- `GOOS=linux`
- `GOARCH=mipsle`
- `GOMIPS=softfloat`
