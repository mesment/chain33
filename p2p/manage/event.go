// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package manage

import (
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
)

var (
	log = log15.New("module", "p2p.manage")
)

/** 处理系统发送至p2p模块的事件, 包含下载, 交易及区块广播等
 * 主要是为了兼容多种类型p2p, 为不同的事件指定路由策略
 */
func (mgr *P2PMgr) handleSysEvent() {

	mgr.Client.Sub("p2p")
	log.Debug("P2PMgr handleSysEvent start")
	for msg := range mgr.Client.Recv() {

		switch msg.Ty {

		case types.EventTxBroadcast, types.EventBlockBroadcast: //广播
			mgr.pub2All(msg)
		case types.EventFetchBlocks, types.EventGetMempool, types.EventFetchBlockHeaders:
			mgr.pub2P2P(msg, mgr.p2pCfg.Types[0])
		case types.EventPeerInfo:
			// 采用默认配置
			p2pTy := mgr.p2pCfg.Types[0]
			req, _ := msg.Data.(*types.P2PGetPeerReq)
			for _, ty := range mgr.p2pCfg.Types {
				if ty == req.GetP2PType() {
					p2pTy = req.GetP2PType()
				}
			}
			mgr.pub2P2P(msg, p2pTy)

		case types.EventGetNetInfo:
			// 采用默认配置
			p2pTy := mgr.p2pCfg.Types[0]
			req, _:= msg.Data.(*types.P2PGetNetInfoReq)
			for _, ty := range mgr.p2pCfg.Types {
				if ty == req.GetP2PType() {
					p2pTy = req.GetP2PType()
				}
			}
			mgr.pub2P2P(msg, p2pTy)

		default:
			log.Warn("unknown msgtype", "msg", msg)
			msg.Reply(mgr.Client.NewMessage("", msg.Ty, types.Reply{Msg: []byte("unknown msgtype")}))
			continue
		}
	}
	log.Debug("P2PMgr handleSysEvent stop")
}

// 处理p2p内部向外发送的消息, 主要是为了兼容多种类型p2p广播消息, 避免重复接交易或者区块
func (mgr *P2PMgr) handleP2PSub() {

	mgr.subChan = mgr.PubSub.Sub("p2p")
	log.Debug("P2PMgr handleP2PSub start")
	//for msg := range mgr.subChan {
	//
	//}

}

// PubBroadCast 兼容多种类型p2p广播消息, 避免重复接交易或者区块
func (mgr *P2PMgr) PubBroadCast(hash string, data interface{}, eventTy int) error {

	exist, _ := mgr.broadcastFilter.ContainsOrAdd(hash, true)
	// eventTy, 交易=1, 区块=54
	log.Debug("PubBroadCast", "eventTy", eventTy, "hash", hash, "exist", exist)
	if exist {
		return nil
	}
	var err error
	if eventTy == types.EventTx {
		err = mgr.Client.Send(mgr.Client.NewMessage("mempool", types.EventTx, data), false)
	} else if eventTy == types.EventBroadcastAddBlock {
		err = mgr.Client.Send(mgr.Client.NewMessage("blockchain", types.EventBroadcastAddBlock, data), false)
	}
	if err != nil {
		log.Error("PubBroadCast", "eventTy", eventTy, "sendMsgErr", err)
	}
	return err
}

//
func (mgr *P2PMgr) pub2All(msg *queue.Message) {

	for _, ty := range mgr.p2pCfg.Types {
		mgr.PubSub.Pub(msg, ty)
	}

}

//
func (mgr *P2PMgr) pub2P2P(msg *queue.Message, p2pType string) {

	mgr.PubSub.Pub(msg, p2pType)
}
