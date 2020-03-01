package protocol

import (
	"github.com/33cn/chain33/p2pnext/protocol/broadcast"
	"github.com/33cn/chain33/p2pnext/protocol/download"
	"github.com/33cn/chain33/p2pnext/protocol/headers"
	"github.com/33cn/chain33/p2pnext/protocol/peer"
	prototypes "github.com/33cn/chain33/p2pnext/protocol/types"
	"github.com/libp2p/go-libp2p-core/protocol"
)

var MsgIDs []protocol.ID

// Init init p2p protocol
func Init(data *prototypes.GlobalData) {
	MsgIDs = append(MsgIDs, peer.PeerInfoReq)
	MsgIDs = append(MsgIDs, broadcast.ID)
	MsgIDs = append(MsgIDs, download.DownloadBlockReq)
	MsgIDs = append(MsgIDs, headers.HeaderInfoReq)
	manager := &prototypes.ProtocolManager{}
	manager.Init(data)
}
