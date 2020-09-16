package test

import (
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer_test/check"
	"github.com/bCoder778/qitmeer_test/conf"
	"github.com/bCoder778/qitmeer_test/rpc"
	"github.com/bCoder778/qitmeer_test/test/node"
)

var Release *node.Node
var Test *node.Node

func TestQitmeer() {
	Release = node.New(rpc.NewClient(&rpc.RpcAuth{
		Host: conf.Setting.ReleaseNode.Host,
		User: conf.Setting.ReleaseNode.User,
		Pwd:  conf.Setting.ReleaseNode.Pass,
	}))

	Test = node.New(rpc.NewClient(&rpc.RpcAuth{
		Host: conf.Setting.TestNode.Host,
		User: conf.Setting.TestNode.User,
		Pwd:  conf.Setting.TestNode.Pass,
	}))

	log.Infof("Start qitmeer test, release=%s, test=%s", Release.Version(), Test.Version())
	order := conf.Setting.Order
	if order == 0 {
		order = Release.BlockCount()
	}
	reBlocks := Release.Sync(order)
	tsBlocks := Test.Sync(order)

	validators, err := check.New(Release.Version(), Test.Version())
	if err != nil {
		log.Errorf("Failed to create check.err=%s", err.Error())
		return
	}
	validators.CheckNode(reBlocks, tsBlocks)
	validators.Close()
	log.Mail("Test Qitmeer Report", validators.SendReport())
}
