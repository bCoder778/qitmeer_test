package node

import (
	"fmt"
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer_test/rpc"
	"strconv"
	"time"
)

type Node struct {
	client  *rpc.Client
	version string
}

func New(client *rpc.Client) *Node {
	info, _ := client.GetNodeInfo()
	return &Node{client: client, version: info.Buildversion}
}

func (n *Node) Sync(lastOrder uint64) chan *rpc.Block {
	blocks := make(chan *rpc.Block, 100)
	var start uint64
	go func() {
		for start <= lastOrder {
			if start%2000 == 0 {
				log.Mail(fmt.Sprintf("Test qitmeer progress %.2f%", start*100/lastOrder))
			}
			block, ok := n.client.GetBlock(start)
			if !ok {
				time.Sleep(time.Second * 10)
			} else {
				if block.Confirmations > 720 {
					color, err := n.client.IsBlue(block.Hash)
					if err != nil {
						time.Sleep(time.Second * 10)
					} else {
						block.IsBlue = color
						blocks <- block
						start++
					}
				} else {
					time.Sleep(time.Second * 10)
				}
			}
		}
		close(blocks)
	}()
	return blocks
}

func (n *Node) BlockCount() uint64 {
	count := n.client.GetBlockCount()
	iCount, _ := strconv.ParseUint(count, 10, 64)
	return iCount
}

func (n *Node) Version() string {
	return n.version
}
