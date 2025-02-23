package main

// . "github.com/stevegt/goadapt"

// XXX probably not this
type Market interface {
	Quote(symbol, denomination []byte) (price float64, err error)
	Order(symbol, denomination []byte, amount float64) (OrderId []byte, err error)
}

// maybe an event-based model instead of a request-response model,
// like the following.  This version requires that the kernel
// (re)parse each received message.
type AgentByte interface {
	Subscribe(prefix []byte) (subId []byte, msgbuf chan []byte, err error)
	Unsubscribe(subId []byte) (err error)
}

// here's the same thing, but where the counterparty parses the
// message before sending it.  This is more efficient, but requires
// that the counterparty be trusted.
type Msg struct {
	// Msg is a COSE-signed CWT
	// XXX this is a placeholder
}

type AgentMsg interface {
	Subscribe(prefix []Msg) (subId []byte, msgbuf chan Msg, err error)
	Unsubscribe(subId []byte) (err error)
}
