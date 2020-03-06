package p2pnext

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"testing"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"

	"crypto/rand"
	l "github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/p2p/manage"
	pmgr "github.com/33cn/chain33/p2p/manage"
	core "github.com/libp2p/go-libp2p-core"
	crypto "github.com/libp2p/go-libp2p-core/crypto"

	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/wallet"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/multiformats/go-multiaddr"

	"github.com/stretchr/testify/assert"
)

func init() {
	l.SetLogLevel("err")
}

func processMsg(q queue.Queue) {
	go func() {
		cfg := q.GetConfig()
		wcli := wallet.New(cfg)
		client := q.Client()
		wcli.SetQueueClient(client)
		//导入种子，解锁钱包
		password := "a12345678"
		seed := "cushion canal bitter result harvest sentence ability time steel basket useful ask depth sorry area course purpose search exile chapter mountain project ranch buffalo"
		saveSeedByPw := &types.SaveSeedByPw{Seed: seed, Passwd: password}
		_, err := wcli.GetAPI().ExecWalletFunc("wallet", "SaveSeed", saveSeedByPw)
		if err != nil {
			return
		}
		walletUnLock := &types.WalletUnLock{
			Passwd:         password,
			Timeout:        0,
			WalletOrTicket: false,
		}

		_, err = wcli.GetAPI().ExecWalletFunc("wallet", "WalletUnLock", walletUnLock)
		if err != nil {
			return
		}
	}()

	go func() {
		blockchainKey := "blockchain"
		client := q.Client()
		client.Sub(blockchainKey)
		for msg := range client.Recv() {
			switch msg.Ty {
			case types.EventGetBlocks:
				if req, ok := msg.GetData().(*types.ReqBlocks); ok {
					if req.Start == 1 {
						msg.Reply(client.NewMessage(blockchainKey, types.EventBlocks, &types.Transaction{}))
					} else {
						msg.Reply(client.NewMessage(blockchainKey, types.EventBlocks, &types.BlockDetails{}))
					}
				} else {
					msg.ReplyErr("Do not support", types.ErrInvalidParam)
				}

			case types.EventGetHeaders:
				if req, ok := msg.GetData().(*types.ReqBlocks); ok {
					if req.Start == 10 {
						msg.Reply(client.NewMessage(blockchainKey, types.EventHeaders, &types.Transaction{}))
					} else {
						msg.Reply(client.NewMessage(blockchainKey, types.EventHeaders, &types.Headers{}))
					}
				} else {
					msg.ReplyErr("Do not support", types.ErrInvalidParam)
				}

			case types.EventGetLastHeader:
				msg.Reply(client.NewMessage("p2p", types.EventHeader, &types.Header{Height: 2019}))
			case types.EventGetBlockHeight:

				msg.Reply(client.NewMessage("p2p", types.EventReplyBlockHeight, &types.ReplyBlockHeight{Height: 2019}))

			}

		}

	}()

	go func() {
		mempoolKey := "mempool"
		client := q.Client()
		client.Sub(mempoolKey)
		for msg := range client.Recv() {
			switch msg.Ty {
			case types.EventGetMempoolSize:
				msg.Reply(client.NewMessage("p2p", types.EventMempoolSize, &types.MempoolSize{Size: 0}))
			}
		}
	}()
}

func NewP2p(cfg *types.Chain33Config) pmgr.IP2P {
	p2pmgr := manage.NewP2PMgr(cfg)
	subCfg := p2pmgr.ChainCfg.GetSubConfig().P2P
	p2p := New(p2pmgr, subCfg[pmgr.DHTTypeName])
	p2p.StartP2P()
	return p2p
}

func testP2PEvent(t *testing.T, qcli queue.Client) {

	msg := qcli.NewMessage("p2p", types.EventBlockBroadcast, &types.Block{})
	assert.True(t, qcli.Send(msg, false) == nil)

	msg = qcli.NewMessage("p2p", types.EventTxBroadcast, &types.Transaction{})
	qcli.Send(msg, false)
	assert.True(t, qcli.Send(msg, false) == nil)

	msg = qcli.NewMessage("p2p", types.EventFetchBlocks, &types.ReqBlocks{})
	qcli.Send(msg, false)
	assert.True(t, qcli.Send(msg, false) == nil)

	msg = qcli.NewMessage("p2p", types.EventGetMempool, nil)
	qcli.Send(msg, false)
	assert.True(t, qcli.Send(msg, false) == nil)

	msg = qcli.NewMessage("p2p", types.EventPeerInfo, nil)
	qcli.Send(msg, false)
	assert.True(t, qcli.Send(msg, false) == nil)

	msg = qcli.NewMessage("p2p", types.EventGetNetInfo, nil)
	qcli.Send(msg, false)
	assert.True(t, qcli.Send(msg, false) == nil)

	msg = qcli.NewMessage("p2p", types.EventFetchBlockHeaders, &types.ReqBlocks{})
	qcli.Send(msg, false)
	assert.True(t, qcli.Send(msg, false) == nil)

}

func testP2PClose(t *testing.T, p2p pmgr.IP2P) {
	p2p.CloseP2P()

}

func testStreamEOFReSet(t *testing.T) {
	r := rand.Reader
	prvKey1, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}
	r = rand.Reader
	prvKey2, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	r = rand.Reader
	prvKey3, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	msgID := "/streamTest"
	h1 := newHost(12345, prvKey1, nil)
	h2 := newHost(12346, prvKey2, nil)
	h3 := newHost(12347, prvKey3, nil)
	h1.SetStreamHandler(protocol.ID(msgID), func(s core.Stream) {
		t.Log("Meow! It worked!")
		var buf []byte
		_, err := s.Read(buf)
		if err != nil {
			t.Log("readStreamErr", err)
		}
		s.Close()
	})

	h2.SetStreamHandler(protocol.ID(msgID), func(s core.Stream) {
		t.Log("H2 Stream! It worked!")
		s.Close()
	})

	h3.SetStreamHandler(protocol.ID(msgID), func(s core.Stream) {
		t.Log("H3 Stream! It worked!")
		s.Conn().Close()
	})

	h2info := peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	}
	err = h1.Connect(context.Background(), h2info)
	assert.NoError(t, err)

	s, err := h1.NewStream(context.Background(), h2.ID(), protocol.ID(msgID))
	assert.NoError(t, err)

	s.Write([]byte("hello"))
	var buf = make([]byte, 128)
	_, err = s.Read(buf)
	assert.True(t, err != nil)
	if err != nil {
		//在stream关闭的时候返回EOF
		t.Log("readStream from H2 Err", err)
		assert.Equal(t, err.Error(), "EOF")
	}

	h3info := peer.AddrInfo{
		ID:    h3.ID(),
		Addrs: h3.Addrs(),
	}
	err = h1.Connect(context.Background(), h3info)
	assert.NoError(t, err)
	s, err = h1.NewStream(context.Background(), h3.ID(), protocol.ID(msgID))
	assert.NoError(t, err)

	s.Write([]byte("hello"))
	_, err = s.Read(buf)
	assert.True(t, err != nil)
	if err != nil {
		//在连接断开的时候，返回 stream reset
		t.Log("readStream from H3 Err", err)
		assert.Equal(t, err.Error(), "stream reset")
	}

}

func TestRelay(t *testing.T)  {
	r := rand.Reader
	prvKey1, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}
	r = rand.Reader
	prvKey2, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	r = rand.Reader
	prvKey3, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	msgID := "/relaymsg"
	var h1options  []libp2p.Option
	var h2options  []libp2p.Option
	var h3options  []libp2p.Option
	h1options = append(h1options,libp2p.Identity(prvKey1),
		libp2p.BandwidthReporter(metrics.NewBandwidthCounter()),
		libp2p.NATPortMap(), libp2p.EnableRelay(circuit.OptDiscovery))
	h2options = append(h2options,libp2p.Identity(prvKey2),
		libp2p.BandwidthReporter(metrics.NewBandwidthCounter()),
		libp2p.NATPortMap(),libp2p.EnableRelay(circuit.OptHop))
	h3options = append(h3options,libp2p.Identity(prvKey3),
		libp2p.BandwidthReporter(metrics.NewBandwidthCounter()),
		libp2p.NATPortMap(),libp2p.EnableRelay())

	h1 := newRelayHost(9980, h1options...)
	h2 := newRelayHost(9981, h2options...)
	h3 := newRelayHost(9982, h3options...)

	h2info := peer.AddrInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	}
	err = h1.Connect(context.Background(), h2info)
	assert.NoError(t, err)

	err = h3.Connect(context.Background(), h2info)
	assert.NoError(t, err)

	h3.SetStreamHandler(protocol.ID(msgID), func(s network.Stream) {
		fmt.Println("Got Relay msg!")
		s.Close()
	})

	_, err = h1.NewStream(context.Background(), h3.ID(), protocol.ID(msgID))
	assert.True(t, err != nil)

	relayaddr, err := multiaddr.NewMultiaddr("/p2p-circuit/ipfs/" + h3.ID().Pretty())
	if err != nil {
		panic(err)
	}
	h3relayInfo := peer.AddrInfo{
		ID:    h3.ID(),
		Addrs: []multiaddr.Multiaddr{relayaddr},
	}
	h1.Network().(*swarm.Swarm).Backoff().Clear(h3.ID())
	err = h1.Connect(context.Background(), h3relayInfo)
	assert.NoError(t, err)
	s, err := h1.NewStream(context.Background(), h3.ID(), protocol.ID(msgID))
	s.Write([]byte("test"))
	assert.NoError(t, err)
	s.Read(make([]byte, 1))

}
func newRelayHost(port int, opts ...libp2p.Option) core.Host {
	m, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		return nil
	}
	optArr := append(opts, libp2p.ListenAddrs(m))
	options := libp2p.ChainOptions(optArr...)
	host, err := libp2p.New(context.Background(), options)
	if err != nil {
		panic(err)
	}
	return host
}
func Test_p2p(t *testing.T) {

	cfg := types.NewChain33Config(types.ReadFile("../cmd/chain33/chain33.test.toml"))
	q := queue.New("channel")
	q.SetConfig(cfg)
	processMsg(q)
	p2p := NewP2p(cfg)
	testP2PEvent(t, q.Client())
	testP2PClose(t, p2p)
	testStreamEOFReSet(t)
}
