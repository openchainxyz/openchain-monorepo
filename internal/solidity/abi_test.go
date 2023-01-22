package solidity

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/openchainxyz/openchainxyz-monorepo/internal/ethclient"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DecodeEventSignature(t *testing.T) {
	for _, tc := range []struct {
		def string
		sig string
	}{
		{`event Transfer(address indexed src, address indexed dst, uint wad)`, `Transfer(address,address,uint256)`},
		{`event Transfer(address indexed, address indexed, uint)`, `Transfer(address,address,uint256)`},
		{`event Transfer(address, address, uint)`, `Transfer(address,address,uint256)`},
		{`Transfer(address indexed src, address indexed dst, uint wad)`, `Transfer(address,address,uint256)`},
		{`Transfer(address, address, uint)`, `Transfer(address,address,uint256)`},
		{`Transfer(address,address,uint)`, `Transfer(address,address,uint256)`},
		{`Transfer(       address       ,       address       ,       uint       )`, `Transfer(address,address,uint256)`},
	} {
		res, err := DecodeEventSignature(tc.def)
		assert.NoError(t, err)

		assert.Equal(t, tc.sig, res.Sig)
	}
}

func Test_VerifySignature(t *testing.T) {
	for _, tc := range []struct {
		sig      string
		expected bool
	}{
		//{``, false},
		//{`()`, false},
		//{`a()`, true},
		//{`$()`, true},
		//{`transfer(address,uint256)`, true},
		//{`transfer(address,uint)`, false},
		//{`transfer(address,uint`, false},
		//{`myFunc((uint256,bool,bool))`, true},
		//{`myFunc((uint256[][][],bool,bool))`, true},
		{`enterArena(uint256[4],address)`, true},
		{`a(,)`, false},
		{`a(()`, false},
		{`a(uint256[[]])`, false},
		{`a(uint256,)`, false},
		{`a(uint256[0])`, false},
	} {
		ok := VerifySignature(tc.sig)
		assert.Equal(t, tc.expected, ok, tc.sig)
	}
}

func Test_AbiDecode(t *testing.T) {
	client, err := ethclient.Dial("https://ethereum-rpc.svc.samczsun.com/")
	assert.NoError(t, err)

	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash("0x5d2813e532a21688bd6bcce9731b8335019c5aab9370a942587a89d23c471e85"))
	assert.NoError(t, err)

	event, err := DecodeEventSignature(`Transfer(address indexed src, address indexed dst, uint wad)`)
	assert.NoError(t, err)

	fmt.Println(event.Sig)

	topics := make(map[string]interface{})
	err = abi.ParseTopicsIntoMap(topics, event.Inputs[:2], receipt.Logs[0].Topics[1:])
	assert.NoError(t, err)

	fmt.Println(topics["src"])

	fmt.Printf("%v\n", event.Inputs)

	assert.Error(t, nil)
}
