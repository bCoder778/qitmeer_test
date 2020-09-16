package check

import (
	"fmt"
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer_test/check/check_db"
	"github.com/bCoder778/qitmeer_test/rpc"
	"os"
	"time"
)

const release_db = "release_db"
const test_db = "test_db"

type Check struct {
	releaseVerify *FeesVerify
	testVerify    *FeesVerify
	stop          chan bool
	errs          []error
	releaseVer    string
	testVer       string
	curBlock      uint64
	start         int64
	ReleaseCount  uint64
	TestCount     uint64
	ReleaseUtxo   uint64
	TestUtxo      uint64
}

func New(releaseVer, testVer string) (*Check, error) {
	releaseVerify, err := NewFeesVerify(release_db)
	if err != nil {
		return nil, fmt.Errorf("create release verify failed!err=%s", err.Error())
	}
	testVerify, err := NewFeesVerify(test_db)
	if err != nil {
		return nil, fmt.Errorf("create test verify failed!err=%s", err.Error())
	}
	return &Check{
		releaseVerify: releaseVerify,
		testVerify:    testVerify,
		stop:          make(chan bool),
		releaseVer:    releaseVer,
		testVer:       testVer,
		errs:          make([]error, 0),
		start:         time.Now().Unix(),
	}, nil
}

func (c *Check) CheckNode(releaseBlocks chan *rpc.Block, testBlocks chan *rpc.Block) {

	defer func() {
		if err := c.VerifyAccount(); err != nil {
			log.Mail(fmt.Sprintf("Verify account failed, %s", err.Error()))
			c.errs = append(c.errs, err)
		}
	}()
	for {
		select {
		case _, _ = <-c.stop:
			return
		default:
			reBlock, ok := <-releaseBlocks
			if !ok {
				return
			}
			tsBlock, ok := <-testBlocks
			if !ok {
				return
			}
			if err := c.VerifyConsistency(reBlock, tsBlock); err != nil {
				log.Mail(fmt.Sprintf("Order %d verification of consistency failed", reBlock.Order), err.Error())
				c.errs = append(c.errs, err)
			}
			c.ReleaseCount++
			if err := c.VerifyFees(reBlock, tsBlock); err != nil {
				log.Mail(fmt.Sprintf("Order %d verification of fees failed", reBlock.Order), err.Error())
				c.errs = append(c.errs, err)
			}
			c.TestCount++
			c.curBlock = reBlock.Order
		}
	}
}

func (c *Check) SendReport() string {
	rs := fmt.Sprintf("Test relase=%s, test=%s use=%ds, blockcount=%d, release-utxo=%d, test-utxo=%d, verify block %d and find %d errors.\n\n\n",
		c.releaseVer, c.testVer, time.Now().Unix()-c.start, c.ReleaseCount, c.ReleaseUtxo, c.TestUtxo, c.curBlock, len(c.errs))
	for _, err := range c.errs {
		rs += err.Error() + "\n"
	}
	return rs
}

func (c *Check) Stop() {
	c.stop <- true
}

func (c *Check) Close() {
	c.testVerify.Close()
	c.releaseVerify.Close()
}

func (c *Check) VerifyConsistency(releaseBlock, testBlock *rpc.Block) error {
	if releaseBlock.Order != testBlock.Order {
		return fmt.Errorf("relase %s block %d, test %s block = %d.", c.releaseVer, releaseBlock.Order, c.testVer, testBlock.Order)
	}
	if releaseBlock.Hash != testBlock.Hash {
		return fmt.Errorf("relase %s block order=%d, hash=%s, test %s block order=%d, hash=%s.",
			c.releaseVer, releaseBlock.Order, releaseBlock.Hash, c.testVer, testBlock.Order, testBlock.Hash)
	}
	if releaseBlock.Txsvalid != testBlock.Txsvalid {
		return fmt.Errorf("block order=%d, relase %s txsvalid=%v, test %s txsvalid=%v.",
			releaseBlock.Order, c.releaseVer, releaseBlock.Txsvalid, c.testVer, testBlock.Txsvalid)
	}
	if releaseBlock.IsBlue != testBlock.IsBlue {
		return fmt.Errorf("block order=%d, relase %s isBlue=%d, test %s isBlue=%d.",
			releaseBlock.Order, c.releaseVer, releaseBlock.IsBlue, c.testVer, testBlock.IsBlue)
	}
	return nil
}

func (c *Check) VerifyFees(releaseBlock, testBlock *rpc.Block) error {
	if err := c.releaseVerify.verify(releaseBlock); err != nil {
		return fmt.Errorf("release %s %s", c.releaseVer, err.Error())
	}
	if err := c.testVerify.verify(testBlock); err != nil {
		return fmt.Errorf("test %s %s", c.testVer, err.Error())
	}
	return nil
}

func (c *Check) Verify(releaseBlock, testBlock *rpc.Block) error {
	if err := c.releaseVerify.verify(releaseBlock); err != nil {
		return fmt.Errorf("release %s %s", c.releaseVer, err.Error())
	}
	if err := c.testVerify.verify(testBlock); err != nil {
		return fmt.Errorf("test %s %s", c.testVer, err.Error())
	}
	return nil
}

func (c *Check) VerifyAccount() error {
	var err error
	c.TestUtxo, err = c.testVerify.SumUTXO()
	if err != nil {
		return fmt.Errorf("test %s sum utxo failed, %s", c.testVer, err.Error())
	}
	c.ReleaseUtxo, err = c.releaseVerify.SumUTXO()
	if err != nil {
		return fmt.Errorf("relesase %s sum utxo failed, %s", c.releaseVer, err.Error())
	}
	correct := (c.TestCount-1)*12000000000 + 6524293004366634
	if c.TestUtxo != correct {
		return fmt.Errorf("test %s sum utxo=%d,blockcount=%d,correct=correct", c.testVer, c.TestUtxo, c.TestCount)
	}
	if c.ReleaseUtxo != correct {
		return fmt.Errorf("release %s sum utxo=%d,blockcount=%d,correct=correct", c.releaseVer, c.ReleaseUtxo, c.ReleaseCount)
	}
	return nil
}

type FeesVerify struct {
	db *check_db.CheckDB
}

func NewFeesVerify(path string) (*FeesVerify, error) {
	if err := os.RemoveAll(path); err != nil {
		return nil, fmt.Errorf("remove releaseDB failed!err=%s", err.Error())
	}
	db, err := check_db.NewCheckDB(path)
	if err != nil {
		return nil, err
	}
	return &FeesVerify{db: db}, nil
}

func (f *FeesVerify) verify(block *rpc.Block) error {
	if _, err := f.checkBlockFee(block); err != nil {
		return err
	}
	f.db.UpdateLastOrder(block.Order)
	return nil
}

func (f *FeesVerify) checkBlockFee(b *rpc.Block) (bool, error) {
	if !b.Txsvalid {
		return true, nil
	}
	err := f.saveVouts(b)
	if err != nil {
		return false, fmt.Errorf("save utxo failed! %s.", err.Error())
	}
	err = f.updateVouts(b)
	if err != nil {
		return false, fmt.Errorf("update utxo failed! %s.", err.Error())
	}

	if b.Order == 0 && b.Id == 0 {
		return true, nil
	}

	var coinbase uint64
	var fee uint64
	for _, tx := range b.Transactions {
		if isCoinBase(&tx) {
			coinbase = tx.Vout[0].Amount
		} else if !tx.Duplicate {
			vinAmount, err := f.sumVin(tx.Vin)
			if err != nil {
				return false, err
			}
			voutAmount := sumVout(tx.Vout)
			fee += vinAmount - voutAmount
		}
	}
	fee += 12000000000
	if coinbase != fee {
		w := &check_db.Wrong{Hash: b.Hash, Order: b.Order, Coinbase: coinbase, CalCoinbase: fee}
		f.db.AddWrong(w)
		return false, fmt.Errorf("find wrong fee block order=%d, hash=%s, coinbase=%d, correct=%d.", w.Order, w.Hash, w.Coinbase, w.CalCoinbase)
	}
	return true, nil
}

func (f *FeesVerify) sumVin(vins []rpc.Vin) (uint64, error) {
	var sum uint64
	for _, vin := range vins {
		amount, err := f.db.GetUTXO(vin.Txid, vin.Vout)
		if err != nil {
			return 0, fmt.Errorf("%s:%d %s.", vin.Txid, vin.Vout, err.Error())
		}
		sum += amount.Amount
	}
	return sum, nil
}

func (f *FeesVerify) saveVouts(b *rpc.Block) error {
	for _, tx := range b.Transactions {
		if !tx.Duplicate {
			for index, vout := range tx.Vout {
				if err := f.db.SaveUTXO(tx.Txid, uint64(index), &check_db.UTXO{Amount: vout.Amount}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (f *FeesVerify) updateVouts(b *rpc.Block) error {
	for _, tx := range b.Transactions {
		if !tx.Duplicate {
			for _, vin := range tx.Vin {
				if vin.Txid != "" {
					if err := f.db.UpdateUTXO(vin.Txid, vin.Vout, tx.Txid); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (f *FeesVerify) SumUTXO() (uint64, error) {
	return f.db.SumUTXO()
}

func (f *FeesVerify) Close() {
	f.db.Close()
}

func sumVout(vouts []rpc.Vout) uint64 {
	var sum uint64
	for _, vout := range vouts {
		sum += vout.Amount
	}
	return sum
}

func isCoinBase(tx *rpc.Transaction) bool {
	if tx != nil && len(tx.Vin) > 0 && tx.Vin[0].Coinbase != "" {
		return true
	}
	return false
}
