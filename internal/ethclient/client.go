package ethclient

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"net/http"
	"strings"
)

type Client struct {
	*ethclient.Client

	C *rpc.Client
}

type transport struct {
	headers map[string]string
	base    http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Add(k, v)
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

func Dial(url string) (*Client, error) {
	var c *rpc.Client
	var err error
	if strings.Contains(url, "samczsun.net") {
		c, err = rpc.DialHTTPWithClient(url, &http.Client{
			Transport: &transport{
				headers: map[string]string{
					"User-Agent": "secret",
				},
			},
		})
	} else {
		c, err = rpc.Dial(url)
	}
	if err != nil {
		return nil, err
	}
	eth := ethclient.NewClient(c)

	return &Client{
		Client: eth,
		C:      c,
	}, nil
}

type Account struct {
	Nonce     *big.Int
	Code      []byte
	Balance   *big.Int
	State     map[common.Hash]common.Hash
	StateDiff map[common.Hash]common.Hash
}

func (ec *Client) SubscribePendingTransactions(ctx context.Context, ch chan<- *types.Transaction) (ethereum.Subscription, error) {
	return ec.C.EthSubscribe(ctx, ch, "newPendingTransactions")
}

func (ec *Client) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	var receipts types.Receipts
	err := ec.C.CallContext(ctx, &receipts, "eth_getReceipts", hash)
	if err != nil {
		return nil, err
	}
	return receipts, nil
}

func (ec *Client) CallContractWithState(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int, overrides map[common.Address]Account) ([]byte, error) {
	var hex hexutil.Bytes
	err := ec.C.CallContext(ctx, &hex, "eth_call", toCallArg(msg), toBlockNumArg(blockNumber), overridesToCallArg(overrides))
	if err != nil {
		return nil, err
	}
	return hex, nil
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}

func overridesToCallArg(overrides map[common.Address]Account) interface{} {
	arg := make(map[string]interface{})
	for k, v := range overrides {
		data := make(map[string]interface{})

		if v.Nonce != nil {
			data["nonce"] = (*hexutil.Big)(v.Nonce)
		}
		if v.Code != nil {
			data["code"] = hexutil.Bytes(v.Code)
		}
		if v.Balance != nil {
			data["balance"] = (*hexutil.Big)(v.Balance)
		}
		if v.State != nil {
			data["state"] = encodeState(v.State)
		}
		if v.StateDiff != nil {
			data["stateDiff"] = encodeState(v.StateDiff)
		}

		arg[k.String()] = data
	}
	return arg
}

func encodeState(state map[common.Hash]common.Hash) interface{} {
	arg := make(map[string]string)
	for k, v := range state {
		arg[k.String()] = v.String()
	}
	return arg
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}

type CallFrame struct {
	Type    string      `json:"type"`
	From    string      `json:"from"`
	To      string      `json:"to,omitempty"`
	Value   string      `json:"value,omitempty"`
	Gas     string      `json:"gas"`
	GasUsed string      `json:"gasUsed"`
	Input   string      `json:"input"`
	Output  string      `json:"output,omitempty"`
	Error   string      `json:"error,omitempty"`
	Calls   []CallFrame `json:"calls,omitempty"`

	LogIndexStart uint `json:"logIndexStart"`
	LogIndexEnd   uint `json:"LogIndexEnd"`
}

type TraceResult struct {
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

func stringPtr(v string) *string {
	return &v
}

func (ec *Client) TraceBlockByHash(ctx context.Context, blockHash common.Hash, config *tracers.TraceConfig) ([]*TraceResult, error) {
	var result []*TraceResult
	config = &tracers.TraceConfig{
		Tracer: stringPtr("customTracer"),
	}
	err := ec.C.CallContext(ctx, &result, "debug_traceBlockByHash", blockHash, config)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ec *Client) TraceTransaction(ctx context.Context, txhash common.Hash, config *tracers.TraceConfig) (json.RawMessage, error) {
	var result json.RawMessage
	err := ec.C.CallContext(ctx, &result, "debug_traceTransaction", txhash, config)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ec *Client) BatchCodeAt(ctx context.Context, addrs []common.Address, blockNumber *big.Int) (map[common.Address][]byte, error) {
	var elems []rpc.BatchElem
	outputs := make([]hexutil.Bytes, len(addrs))
	for i, addr := range addrs {
		elems = append(elems, rpc.BatchElem{
			Method: "eth_getCode",
			Args:   []interface{}{addr, toBlockNumArg(blockNumber)},
			Result: &outputs[i],
			Error:  nil,
		})
	}
	err := ec.C.BatchCallContext(ctx, elems)
	if err != nil {
		return nil, err
	}

	result := make(map[common.Address][]byte)
	for i, addr := range addrs {
		if elems[i].Error != nil {
			return nil, err
		}
		result[addr] = outputs[i]
	}
	return result, nil
}

func (ec *Client) TransactionReceiptsInBlock(ctx context.Context, block *types.Block) ([]*types.Receipt, error) {
	results := make([]*types.Receipt, block.Transactions().Len())

	var elems []rpc.BatchElem
	for i := range results {
		results[i] = new(types.Receipt)
		elems = append(elems, rpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []interface{}{block.Transactions()[i].Hash()},
			Result: results[i],
			Error:  nil,
		})
	}
	err := ec.C.BatchCallContext(ctx, elems)
	if err != nil {
		return nil, err
	}

	for i := range results {
		if elems[i].Error != nil {
			return nil, elems[i].Error
		}
	}
	return results, nil
}
