package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/params"
)

func getChainConfig(chainID uint64) (*params.ChainConfig, error) {
	switch chainID {
	case 0, params.MainnetChainConfig.ChainID.Uint64():
		return params.MainnetChainConfig, nil
	case params.SepoliaChainConfig.ChainID.Uint64():
		return params.SepoliaChainConfig, nil
	case params.HoodiChainConfig.ChainID.Uint64():
		return params.HoodiChainConfig, nil
	default:
		return nil, fmt.Errorf("unsupported chain ID: %d", chainID)
	}
}

