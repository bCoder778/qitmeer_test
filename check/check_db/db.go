package check_db

import (
	"encoding/json"
	"fmt"
	"github.com/bCoder778/qitmeer_test/db/base"
	"github.com/bCoder778/qitmeer_test/encode"
)

const (
	block_bucket  = "block_bucket"
	tx_bucket     = "tx_bucket"
	result_bucket = "result_bucket"
)

type CheckDB struct {
	base *base.Base
}

func NewCheckDB(path string) (*CheckDB, error) {
	base, err := base.Open(path)
	if err != nil {
		return nil, err
	}
	return &CheckDB{base}, nil
}

func (c *CheckDB) Close() {
	c.base.Close()
}

func (c *CheckDB) LastBlockOrder() uint64 {
	bytes, err := c.base.GetFromBucket(block_bucket, []byte(block_bucket))
	if err != nil {
		return 0
	}
	return encode.BytesToUint64(bytes)
}

func (c *CheckDB) UpdateLastOrder(order uint64) {
	c.base.PutInBucket(block_bucket, []byte(block_bucket), encode.Uint64ToBytes(order))
}

func (c *CheckDB) AddWrong(w *Wrong) {
	bytes, _ := w.Bytes()
	c.base.PutInBucket(result_bucket, []byte(w.Hash), bytes)
}

func (c *CheckDB) WrongList() []Wrong {
	rs := c.base.Foreach(result_bucket)
	wrongs := make([]Wrong, 0)
	for _, value := range rs {
		w, _ := BytesToWrong(value)
		wrongs = append(wrongs, *w)
	}
	return wrongs
}

func (c *CheckDB) GetUTXO(txId string, index uint64) (*UTXO, error) {
	bytes, err := c.base.GetFromBucket(tx_bucket, []byte(getOutKey(txId, index)))
	if err != nil {
		return nil, err
	}
	var utxo *UTXO
	err = json.Unmarshal(bytes, &utxo)
	if err != nil {
		return nil, err
	}
	return utxo, nil
}

func (c *CheckDB) SaveUTXO(txId string, index uint64, uxto *UTXO) error {
	bytes, err := json.Marshal(uxto)
	if err != nil {
		return err
	}
	return c.base.PutInBucket(tx_bucket, []byte(getOutKey(txId, index)), bytes)
}

func (c *CheckDB) UpdateUTXO(txId string, index uint64, spent string) error {
	uxto, err := c.GetUTXO(txId, index)
	if err != nil {
		return err
	}
	uxto.Spent = spent
	err = c.SaveUTXO(txId, index, uxto)
	if err != nil {
		return err
	}
	return nil
}

func (c *CheckDB) SumUTXO() (uint64, error) {
	var sum uint64
	iter := c.base.Iter(tx_bucket)
	defer iter.Release()

	// Iter will affect RLP decoding and reallocate memory to value
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		var utxo *UTXO
		err := json.Unmarshal(value, &utxo)
		if err != nil {
			return 0, err
		}
		if utxo.Spent == "" {
			sum += utxo.Amount
		}
	}
	return sum, nil
}

type UTXO struct {
	Amount uint64
	Spent  string
}

type Wrong struct {
	Order       uint64
	Hash        string
	Coinbase    uint64
	CalCoinbase uint64
}

func (w *Wrong) Bytes() ([]byte, error) {
	return json.Marshal(w)
}

func BytesToWrong(bytes []byte) (*Wrong, error) {
	var w *Wrong
	err := json.Unmarshal(bytes, &w)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func getOutKey(txId string, idx interface{}) string {
	return fmt.Sprintf("%s-%d", txId, idx)
}
