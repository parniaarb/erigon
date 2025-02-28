// Copyright 2021 The go-ethereum Authors
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

package misc

import (
	"fmt"
	"math/big"

	"github.com/ledgerwatch/erigon-lib/chain"
	"github.com/ledgerwatch/erigon-lib/common"
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/polygon/bor/borcfg"

	"github.com/ledgerwatch/erigon/common/math"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/ledgerwatch/erigon/params"
)

// VerifyEip1559Header verifies some header attributes which were changed in EIP-1559,
// - gas limit check
// - basefee check
func VerifyEip1559Header(config *chain.Config, parent, header *types.Header, skipGasLimit bool) error {
	if !skipGasLimit {
		// Verify that the gas limit remains within allowed bounds
		parentGasLimit := parent.GasLimit
		if !config.IsLondon(parent.Number.Uint64()) {
			parentGasLimit = parent.GasLimit * params.ElasticityMultiplier
		}
		if err := VerifyGaslimit(parentGasLimit, header.GasLimit); err != nil {
			return err
		}
	}
	// Verify the header is not malformed
	if header.BaseFee == nil {
		return fmt.Errorf("header is missing baseFee")
	}
	// Verify the baseFee is correct based on the parent header.
	expectedBaseFee := CalcBaseFee(config, parent)
	if header.BaseFee.Cmp(expectedBaseFee) != 0 {
		return fmt.Errorf("invalid baseFee: have %s, want %s, parentBaseFee %s, parentGasUsed %d",
			header.BaseFee, expectedBaseFee, parent.BaseFee, parent.GasUsed)
	}
	return nil
}

var Eip1559FeeCalculator eip1559Calculator

type eip1559Calculator struct{}

func (f eip1559Calculator) CurrentFees(chainConfig *chain.Config, db kv.Getter) (baseFee uint64, blobFee uint64, minBlobGasPrice uint64, err error) {
	hash := rawdb.ReadHeadHeaderHash(db)

	if hash == (libcommon.Hash{}) {
		return 0, 0, 0, fmt.Errorf("can't get head header hash")
	}

	currentHeader, err := rawdb.ReadHeaderByHash(db, hash)

	if err != nil {
		return 0, 0, 0, err
	}

	if chainConfig != nil {
		if currentHeader.BaseFee != nil {
			baseFee = CalcBaseFee(chainConfig, currentHeader).Uint64()
		}

		if currentHeader.ExcessBlobGas != nil {
			excessBlobGas := CalcExcessBlobGas(chainConfig, currentHeader)
			b, err := GetBlobGasPrice(chainConfig, excessBlobGas)
			if err == nil {
				blobFee = b.Uint64()
			}
		}
	}

	minBlobGasPrice = chainConfig.GetMinBlobGasPrice()

	return baseFee, blobFee, minBlobGasPrice, nil
}

// CalcBaseFee calculates the basefee of the header.
func CalcBaseFee(config *chain.Config, parent *types.Header) *big.Int {
	// If the current block is the first EIP-1559 block, return the InitialBaseFee.
	if !config.IsLondon(parent.Number.Uint64()) {
		return new(big.Int).SetUint64(params.InitialBaseFee)
	}

	var (
		parentGasTarget          = parent.GasLimit / params.ElasticityMultiplier
		parentGasTargetBig       = new(big.Int).SetUint64(parentGasTarget)
		baseFeeChangeDenominator = new(big.Int).SetUint64(getBaseFeeChangeDenominator(config.Bor, parent.Number.Uint64()))
	)
	// If the parent gasUsed is the same as the target, the baseFee remains unchanged.
	if parent.GasUsed == parentGasTarget {
		return new(big.Int).Set(parent.BaseFee)
	}
	if parent.GasUsed > parentGasTarget {
		// If the parent block used more gas than its target, the baseFee should increase.
		gasUsedDelta := new(big.Int).SetUint64(parent.GasUsed - parentGasTarget)
		x := new(big.Int).Mul(parent.BaseFee, gasUsedDelta)
		y := x.Div(x, parentGasTargetBig)
		baseFeeDelta := math.BigMax(
			x.Div(y, baseFeeChangeDenominator),
			common.Big1,
		)

		return x.Add(parent.BaseFee, baseFeeDelta)
	} else {
		// Otherwise if the parent block used less gas than its target, the baseFee should decrease.
		gasUsedDelta := new(big.Int).SetUint64(parentGasTarget - parent.GasUsed)
		x := new(big.Int).Mul(parent.BaseFee, gasUsedDelta)
		y := x.Div(x, parentGasTargetBig)
		baseFeeDelta := x.Div(y, baseFeeChangeDenominator)

		return math.BigMax(
			x.Sub(parent.BaseFee, baseFeeDelta),
			common.Big0,
		)
	}
}

func getBaseFeeChangeDenominator(borConfig chain.BorConfig, number uint64) uint64 {
	// If we're running bor based chain post delhi hardfork, return the new value
	if borConfig, ok := borConfig.(*borcfg.BorConfig); ok && borConfig.IsDelhi(number) {
		return params.BaseFeeChangeDenominatorPostDelhi
	}

	// Return the original once for other chains and pre-fork cases
	return params.BaseFeeChangeDenominator
}
