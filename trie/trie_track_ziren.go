// Copyright 2022 The go-ethereum Authors
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

//go:build ziren

package trie

// trackResolved records nothing: the original values of resolved nodes are
// read by Commit, for the origins of deleted nodes, and by witness
// generation, and a ziren guest hashes a trie but never commits it or builds
// a witness, while recording every resolved node copies its path into a
// string and inserts it into a map, on 32-bit MIPS about two percent of a
// block. The tracer stays empty, so Commit under this tag reports no
// origins.
func (t *Trie) trackResolved(path, blob []byte) {}
