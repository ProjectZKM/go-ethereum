module github.com/ethereum/go-ethereum/cmd/keeper-subblock

go 1.24.0

require (
	github.com/ProjectZKM/Ziren/crates/go-runtime/zkvm_runtime v0.0.0-20251001021608-1fe7b43fc4d6
	github.com/ethereum/go-ethereum v0.0.0-00010101000000-000000000000
)

replace github.com/ethereum/go-ethereum => ../../

// Keep the same local override pattern as cmd/keeper.
replace github.com/ProjectZKM/Ziren/crates/go-runtime/zkvm_runtime => /data/stephen/Ziren/crates/go-runtime/zkvm_runtime

