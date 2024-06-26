package origins

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"

	pkgEthereum "github.com/toknowwhy/theunit-oracle/pkg/ethereum"
)

// The three values that can be queried:
//
// - PAIR_PRICE: the price of the tokens in the Pool, expressed as the price of the second token in units of the
//   first token. For example, if token A is worth $2, and token B is worth $4, the pair price will be 2.0.
//   Note that the price is computed *including* the tokens decimals. This means that the pair price of a Pool with
//   DAI and USDC will be close to 1.0, despite DAI having 18 decimals and USDC 6.
//
// - BPT_PRICE: the price of the Pool share token (BPT), in units of the first token.
//   Note that the price is computed *including* the tokens decimals. This means that the BPT price of a Pool with
//   USDC in which BPT is worth $5 will be 5.0, despite the BPT having 18 decimals and USDC 6.
//
// - INVARIANT: the value of the Pool's invariant, which serves as a measure of its liquidity.
// enum Variable { PAIR_PRICE, BPT_PRICE, INVARIANT }

//go:embed balancerv2_abi.json
var balancerV2PoolABI string

// TODO: should be configurable
const balancerV2Denominator = 1e18

type BalancerV2 struct {
	ethClient         pkgEthereum.Client
	ContractAddresses ContractAddresses
	abi               abi.ABI
	variable          byte
}

func NewBalancerV2(cli pkgEthereum.Client, addrs ContractAddresses) (*BalancerV2, error) {
	a, err := abi.JSON(strings.NewReader(balancerV2PoolABI))
	if err != nil {
		return nil, err
	}
	return &BalancerV2{
		ethClient:         cli,
		ContractAddresses: addrs,
		abi:               a,
		variable:          0, // PAIR_PRICE
	}, nil
}

func (s BalancerV2) PullPrices(pairs []Pair) []FetchResult {
	return callSinglePairOrigin(&s, pairs)
}

func (s BalancerV2) callOne(pair Pair) (*Price, error) {
	contract, inverted, err := s.ContractAddresses.AddressByPair(pair)
	if err != nil {
		return nil, err
	}
	if inverted {
		return nil, fmt.Errorf("cannot use inverted pair to retrieve price: %s", pair.String())
	}

	var callData []byte
	callData, err = s.abi.Pack("getLatest", s.variable)
	if err != nil {
		return nil, fmt.Errorf("failed to pack contract args for pair %s: %w", pair.String(), err)
	}

	resp, err := s.ethClient.Call(context.Background(), pkgEthereum.Call{Address: contract, Data: callData})
	if err != nil {
		return nil, err
	}
	bn := new(big.Int).SetBytes(resp)
	price, _ := new(big.Float).Quo(new(big.Float).SetInt(bn), new(big.Float).SetUint64(balancerV2Denominator)).Float64()

	return &Price{
		Pair:      pair,
		Price:     price,
		Timestamp: time.Now(),
	}, nil
}
