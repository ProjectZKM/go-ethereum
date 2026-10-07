// Copyright 2024 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package stateless

import (
	"unsafe"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
)

// MakeHashDB imports tries, codes and block hashes from a witness into a new
// hash-based memory db. We could eventually rewrite this into a pathdb, but
// simple is better for now.
//
// Note, this hashdb approach is quite strictly self-validating:
//   - Headers are persisted keyed by hash, so blockhash will error on junk
//   - Codes are persisted keyed by hash, so bytecode lookup will error on junk
//   - Trie nodes are persisted keyed by hash, so trie expansion will error on junk
//
// Acceleration structures built would need to explicitly validate the witness.
func (w *Witness) MakeHashDB() ethdb.Database {
	memdb := rawdb.NewMemoryDatabase()

	// Inject all the "block hashes" (i.e. headers) into the ephemeral database
	for _, header := range w.Headers {
		rawdb.WriteHeader(memdb, header)
	}
	// Inject all the bytecodes and MPT trie nodes into the ephemeral database.
	// Each blob is read in place from its string: hashing does not write it,
	// and the database stores a copy.
	for code := range w.Codes {
		blob := unsafe.Slice(unsafe.StringData(code), len(code))
		rawdb.WriteCode(memdb, crypto.Keccak256Hash(blob), blob)
	}
	for node := range w.State {
		blob := unsafe.Slice(unsafe.StringData(node), len(node))
		rawdb.WriteLegacyTrieNode(memdb, crypto.Keccak256Hash(blob), blob)
	}
	return memdb
}
