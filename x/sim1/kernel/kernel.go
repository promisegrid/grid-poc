package kernel

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"sim1/wire"

	"github.com/ipfs/go-cid"
)

type Kernel struct {
	mu            sync.RWMutex
	subscriptions map[string]func(wire.Message)
	listener      net.Listener
	peerAddr      string
	conn          net.Conn
	connMu        sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewKernel() *Kernel {
	ctx, cancel := context.WithCancel(context.Background())
	return &Kernel{
		subscriptions: make(map[string]func(wire.Message)),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (k *Kernel) Start(port int) error {
	var err error
	k.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	go k.acceptConnections()
	go k.maintainOutgoingConnection()
	return nil
}

func (k *Kernel) acceptConnections() {
	for {
		conn, err := k.listener.Accept()
		if err != nil {
			select {
			case <-k.ctx.Done():
				return
			default:
				log.Printf("accept error: %v", err)
			}
			continue
		}
		go k.handleConnection(conn)
	}
}

func (k *Kernel) maintainOutgoingConnection() {
	for {
		select {
		case <-k.ctx.Done():
			return
		default:
			if k.peerAddr == "" {
				time.Sleep(1 * time.Second)
				continue
			}

			k.connMu.Lock()
			if k.conn == nil {
				conn, err := net.Dial("tcp", k.peerAddr)
				if err != nil {
					log.Printf("dial error: %v", err)
					k.connMu.Unlock()
					time.Sleep(2 * time.Second)
					continue
				}
				k.conn = conn
				log.Printf("connected to peer %s", k.peerAddr)
				go k.handleConnection(conn)
			}
			k.connMu.Unlock()
			time.Sleep(5 * time.Second)
		}
	}
}

func (k *Kernel) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Set this connection as active
	k.connMu.Lock()
	if k.conn != nil {
		k.conn.Close()
	}
	k.conn = conn
	k.connMu.Unlock()

	for {
		var msg wire.Message
		dec := wire.Dm.NewDecoder(conn)
		err := dec.Decode(&msg)
		if err != nil {
			log.Printf("decode error: %v", err)
			return
		}

		protocolCid, err := cid.Cast(msg.Protocol)
		if err != nil {
			log.Printf("invalid protocol CID: %v", err)
			continue
		}

		k.mu.RLock()
		handler, exists := k.subscriptions[protocolCid.String()]
		k.mu.RUnlock()

		if exists {
			handler(msg)
		}
	}
}

func (k *Kernel) Publish(msg wire.Message) error {
	k.connMu.Lock()
	defer k.connMu.Unlock()

	if k.conn == nil {
		return fmt.Errorf("no active connection to peer")
	}

	enc := wire.Em.NewEncoder(k.conn)
	return enc.Encode(msg)
}

func (k *Kernel) Subscribe(protocol cid.Cid, handler func(wire.Message)) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.subscriptions[protocol.String()] = handler
}

func (k *Kernel) Unsubscribe(protocol cid.Cid) {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.subscriptions, protocol.String())
}

func (k *Kernel) SetPeer(addr string) {
	k.peerAddr = addr
}

func (k *Kernel) Stop() {
	k.cancel()
	k.connMu.Lock()
	if k.conn != nil {
		k.conn.Close()
	}
	k.connMu.Unlock()
	if k.listener != nil {
		k.listener.Close()
	}
}
