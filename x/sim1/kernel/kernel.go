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

// Agent defines the interface that each agent must implement.
type Agent interface {
	Run(context.Context)
	Stop()
}

type Kernel struct {
	mu            sync.RWMutex
	subscriptions map[string]func(wire.Message)
	listener      net.Listener
	peerAddr      string
	conns         map[string]net.Conn
	connMu        sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
	agents        []Agent
}

func NewKernel() *Kernel {
	ctx, cancel := context.WithCancel(context.Background())
	return &Kernel{
		subscriptions: make(map[string]func(wire.Message)),
		conns:         make(map[string]net.Conn),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (k *Kernel) Start(port int) error {
	var err error
	if port > 0 {
		k.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return fmt.Errorf("failed to listen: %v", err)
		}
		log.Printf("listening on port %d", port)
		go k.acceptConnections()
	} else {
		log.Println("not listening for incoming connections")
		go k.maintainOutgoingConnection()
	}
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
		log.Printf("accepted connection from %s", conn.RemoteAddr())
		k.connMu.Lock()
		k.conns[conn.RemoteAddr().String()] = conn
		k.connMu.Unlock()
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
			_, exists := k.conns[k.peerAddr]
			k.connMu.Unlock()
			if !exists {
				conn, err := net.Dial("tcp", k.peerAddr)
				if err != nil {
					log.Printf("dial error: %v", err)
					time.Sleep(2 * time.Second)
					continue
				}
				k.connMu.Lock()
				k.conns[k.peerAddr] = conn
				k.connMu.Unlock()
				log.Printf("connected to peer %s", k.peerAddr)
				go k.handleConnection(conn)
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func (k *Kernel) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		k.connMu.Lock()
		delete(k.conns, conn.RemoteAddr().String())
		k.connMu.Unlock()
	}()
	dec := wire.Dm.NewDecoder(conn)
	for {
		var msg wire.Message
		err := dec.Decode(&msg)
		if err != nil {
			log.Printf("decode error from %s: %v", conn.RemoteAddr(), err)
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

func (k *Kernel) Send(msg wire.Message) error {
	k.connMu.Lock()
	defer k.connMu.Unlock()

	var firstErr error
	for addr, conn := range k.conns {
		enc := wire.Em.NewEncoder(conn)
		err := enc.Encode(msg)
		if err != nil {
			log.Printf("encode error on conn %s: %v", addr, err)
			conn.Close()
			delete(k.conns, addr)
			if firstErr == nil {
				firstErr = fmt.Errorf("failed to encode message on %s: %v", addr, err)
			}
		}
	}
	if firstErr != nil {
		return firstErr
	}
	return nil
}

func (k *Kernel) Register(protocol cid.Cid, handler func(wire.Message)) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.subscriptions[protocol.String()] = handler
}

func (k *Kernel) Deregister(protocol cid.Cid) {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.subscriptions, protocol.String())
}

func (k *Kernel) SetPeer(addr string) {
	k.peerAddr = addr
}

func (k *Kernel) AddAgent(a Agent) {
	k.mu.Lock()
	k.agents = append(k.agents, a)
	k.mu.Unlock()
	go a.Run(k.ctx)
}

func (k *Kernel) Stop() {
	k.cancel()
	k.connMu.Lock()
	for addr, conn := range k.conns {
		conn.Close()
		delete(k.conns, addr)
	}
	k.connMu.Unlock()
	if k.listener != nil {
		k.listener.Close()
	}
}
