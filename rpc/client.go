package rpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/bCoder778/log"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type RpcAuth struct {
	Host string `toml:"host"`
	User string `toml:"user"`
	Pwd  string `toml:"pwd"`
}

type Client struct {
	rpcAuth *RpcAuth
}

func NewClient(auth *RpcAuth) *Client {
	return &Client{auth}
}

func (c *Client) GetBlock(h uint64) (*Block, bool) {
	params := []interface{}{h, true}
	resp := NewReqeust(params).SetMethod("getBlockByOrder").call(c.rpcAuth)
	blk := new(Block)
	if resp.Error != nil {
		return blk, false
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		log.Error(err.Error())
		return blk, false
	}
	return blk, true
}

func (c *Client) GetBlockByHash(hash string) (*Block, bool) {
	params := []interface{}{hash, true}
	resp := NewReqeust(params).SetMethod("getBlock").call(c.rpcAuth)
	blk := new(Block)
	if resp.Error != nil {
		return blk, false
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		return blk, false
	}
	return blk, true
}

func (c *Client) GetBlockCount() string {
	var params []interface{}
	resp := NewReqeust(params).SetMethod("getBlockCount").call(c.rpcAuth)
	if resp.Error != nil {
		return "-1"
	}
	return string(resp.Result)
}

func (c *Client) GetMainChainHeight() string {
	var params []interface{}
	resp := NewReqeust(params).SetMethod("getMainChainHeight").call(c.rpcAuth)
	if resp.Error != nil {
		return "-1"
	}
	return string(resp.Result)
}

func (c *Client) SendTransaction(tx string) (string, bool) {
	params := []interface{}{strings.Trim(tx, "\n"), false}
	resp := NewReqeust(params).SetMethod("sendRawTransaction").call(c.rpcAuth)
	if resp.Error != nil {
		return resp.Error.Message, false
	}
	txId := string(resp.Result)
	return txId, true
}

func (c *Client) GetTransaction(txid string) (*Transaction, error) {
	params := []interface{}{txid, true}
	resp := NewReqeust(params).SetMethod("getRawTransaction").call(c.rpcAuth)
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs *Transaction
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) CreateTransaction(inputs []TransactionInput, amounts Amounts) (string, error) {
	jsonInput, err := json.Marshal(inputs)
	if err != nil {
		return "", err
	}
	jsonAmount, err := json.Marshal(amounts)
	if err != nil {
		return "", err
	}
	params := []interface{}{json.RawMessage(jsonInput), json.RawMessage(jsonAmount)}
	resp := NewReqeust(params).SetMethod("createRawTransaction").call(c.rpcAuth)
	if resp.Error != nil {
		return "", errors.New(resp.Error.Message)
	}
	encode := string(resp.Result)
	return encode, nil
}

func (c *Client) GetMemoryPool() ([]string, error) {
	params := []interface{}{"", false}
	resp := NewReqeust(params).SetMethod("getMempool").call(c.rpcAuth)
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs []string
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) GetPeerInfo() ([]PeerInfo, error) {
	var params []interface{}
	resp := NewReqeust(params).SetMethod("getPeerInfo").call(c.rpcAuth)
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs []PeerInfo
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) GetBlockById(id uint64) (*Block, bool) {
	params := []interface{}{id, true}
	resp := NewReqeust(params).SetMethod("getBlockByID").call(c.rpcAuth)
	blk := new(Block)
	if resp.Error != nil {
		return blk, false
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		log.Error(err.Error())
		return blk, false
	}
	return blk, true
}

func (c *Client) GetNodeInfo() (*NodeInfo, error) {
	params := []interface{}{}
	resp := NewReqeust(params).SetMethod("getNodeInfo").call(c.rpcAuth)
	nodeInfo := new(NodeInfo)
	if resp.Error != nil {
		return nodeInfo, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, nodeInfo); err != nil {
		log.Error(err.Error())
		return nodeInfo, err
	}
	return nodeInfo, nil
}

func (c *Client) IsBlue(hash string) (int, error) {
	params := []interface{}{hash}
	resp := NewReqeust(params).SetMethod("isBlue").call(c.rpcAuth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}
	state, err := strconv.Atoi(string(resp.Result))
	if err != nil {
		return 0, err
	}
	return state, nil
}

func (c *Client) GetFees(hash string) (uint64, error) {
	params := []interface{}{hash}
	resp := NewReqeust(params).SetMethod("getFees").call(c.rpcAuth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}

	return strconv.ParseUint(string(resp.Result), 10, 64)
}

func (req *ClientRequest) call(auth *RpcAuth) *ClientResponse {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	//convert struct to []byte
	marshaledData, err := json.Marshal(req)
	if err != nil {
		log.Error(err.Error())
	}
	//log.Debugf("rpc call starting, Host:%s, params:%s", cfg.Host, marshaledData)

	httpRequest, err :=
		http.NewRequest(http.MethodPost, auth.Host, bytes.NewReader(marshaledData))
	if err != nil {
		log.Error(err.Error())
	}

	if httpRequest == nil {
		log.Error("the httpRequest is nil")
		return &ClientResponse{Error: &Error{Message: "the httpRequest is nil"}}
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.SetBasicAuth(auth.User, auth.Pwd)
	//log.Debugf("u:%s;p:%s", cfg.User, cfg.Pwd)

	response, err := client.Do(httpRequest)
	if err != nil {
		log.Error(err.Error())
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	body := response.Body

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		log.Error("io read error", err.Error())
	}

	//log.Info("rpc call successful! ", string(bodyBytes))

	resp := &ClientResponse{}
	//convert []byte to struct
	if err := json.Unmarshal(bodyBytes, resp); err != nil {
		log.Errorf("json unmarshal failed; value:%s; error:%s", string(bodyBytes), err.Error())
	}

	err = response.Body.Close()
	if err != nil {
		log.Error(err.Error())
	}

	if resp.Error != nil {
		//log.Fail(resp.Error.Message)
	}
	return resp
}
