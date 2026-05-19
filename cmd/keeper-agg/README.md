# keeper-agg (zkVM guest)

`keeper-agg` is a go-ethereum zkVM guest program that **aggregates and constrains** a sequence of
subblock results produced by `keeper-subblock`.

It is designed to be driven by a host-side pipeline that:
1) Computes subblock boundaries (matching the Rust pipeline semantics),
2) Executes/proves each `keeper-subblock` segment,
3) Feeds the collected subblock outputs and, in proof mode, the subblock proof stream into
   `keeper-agg`.

## Goal

- Check basic subblock sequencing constraints (monotonic `tx_end`, coverage of the full tx list).
- In proof mode, verify each `keeper-subblock` proof and bind its public value to the corresponding
  `SubblockOutput`.
- Constrain the *final* execution result by checking that the last subblock roots match the block
  header roots (`stateRoot` and `receiptsRoot`).

## Input (stdin)

With the `ziren` build tag enabled, the program reads four values from stdin via
`zkvm_runtime.Read[T]` (Ziren go-runtime "bincode-like" encoding compatible with Rust `ZKMStdin::write`):

```go
publicValues := zkruntime.Read[[][]byte]()
subblockVK := zkruntime.Read[[8]uint32]()
deferredProofsDigest := zkruntime.Read[[8]uint32]()
payload := zkruntime.Read[AggregationPayloadBytes]()
```

`publicValues` contains each `keeper-subblock` proof public value in subblock order. In execute-only
mode it may be empty, which skips deferred proof verification. In proof mode, the host must also
write the compressed subblock proofs to the zkVM proof stream in the same order.

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
- When `publicValues` is non-empty, `len(publicValues)` must equal `len(Subblocks)`.
- `subblockVK` is the `[8]uint32` verifier-key digest for the `keeper-subblock` program.
- `deferredProofsDigest` must match the digest accumulated by the host for the deferred proof stream.

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

## What Is Verified

Current implementation verifies:
- If `publicValues` is present, each subblock proof is verified with `subblockVK` and the SHA-256
  digest of that public value.
- Each verified public value deserializes to `SubblockOutput` and exactly equals the corresponding
  `Subblocks[i]`.
- `tx_end` is non-decreasing and within `[0, total_txs]`
- The last subblock reaches the end (`tx_end == total_txs`)
- The last subblock `(stateRoot, receiptRoot)` matches the block header `(Root, ReceiptHash)`

Not verified:
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
