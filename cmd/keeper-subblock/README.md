# keeper-subblock (zkVM guest)

`keeper-subblock` is a go-ethereum zkVM guest program for **subblock (segmented) execution**.
It is intended to be used by a host-side subblock/aggregation pipeline (similar in shape to the
Rust `reth-subblock`/`reth-agg` flow).

## Goal

- Execute a block transaction prefix `[0:tx_end)` **statelessly** inside the zkVM (using witness data
  produced by `debug_executionWitness`).
- Commit a compact subblock result as public values so the host can aggregate and further constrain
  the full-block execution.

## Input (stdin)

With the `ziren` build tag enabled, the program reads a single struct from stdin via
`zkvm_runtime.Read[T]`.

The encoding is the Ziren go-runtime "bincode-like" encoding and is compatible with the Rust side
when the host writes the same struct with `ZKMStdin::write` (which uses `bincode`).

```go
type SubblockPayloadBytes struct {
    ChainID    uint64
    BlockRLP   []byte // raw block RLP bytes (from debug_getRawBlock)
    WitnessRLP []byte // witness RLP bytes (from debug_executionWitness)
    TxEnd      uint64 // exclusive end index of the tx prefix to execute
}
```

Notes:
- `BlockRLP`/`WitnessRLP` should match the bytes used by `cmd/keeper` payload generation.
- `TxEnd` is chosen by the host to match the Rust pipeline boundary semantics (typically derived
  from `SUBBLOCK_GAS_LIMIT` and receipts `cumulativeGasUsed`).

## Output (public values)

The program commits one struct via `zkvm_runtime.Commit(SubblockOutput)`:

```go
type SubblockOutput struct {
    TxEnd           uint64
    GasUsed         uint64
    StateRoot       [32]byte
    ReceiptRoot     [32]byte
    ExecutedTxCount uint64
}
```

`StateRoot`/`ReceiptRoot` are the roots after executing the prefix subblock, not the final full-block
roots.

## Build

This program is meant for zkVM builds:

```bash
cd cmd/keeper-agg
GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -tags "ziren" ./...
```

Environment (typically provided by the host build script / overlay):
- `GOOS=linux`
- `GOARCH=mipsle`
- `GOMIPS=softfloat`

## Important limitations

Today `debug_executionWitness` returns a witness for the **full block** only.
To enable a subblock pipeline without changing the RPC API, this program executes a transaction
prefix while still using the full-block witness.

Consequences:
- Witness payloads can be larger than necessary for a subblock.
- A future improvement would be adding "range witness" support on the node side so each subblock
  can use a minimal witness.
