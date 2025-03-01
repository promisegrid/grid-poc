package main

// Market interface represents a trading market with capabilities to
// fetch quotes and execute orders. It uses raw []byte for symbols and
// denominations which may be beneficial for low-level manipulations.
type Market interface {
	Quote(symbol, denomination []byte) (price float64, err error)
	Order(symbol, denomination []byte, amount float64) (OrderId []byte, err error)
}

// AgentByte interface represents an event-based model where the
// message content is handled as raw bytes. This offers flexibility
// and universal compatibility but comes with the challenges of
// ambiguity, security, and duplicated parsing.
type AgentByte interface {
	Subscribe(prefix []byte) (subId []byte, msgbuf chan []byte, err error)
	Unsubscribe(subId []byte) (err error)
}

// Msg type is intended to represent a structured message, such as a
// COSE-signed CWT. This structure is a placeholder and should be
// expanded to include actual message fields and validations.
type Msg struct {
	// Msg is a COSE-signed CWT
	// XXX this is a placeholder
}

// AgentMsg interface uses a pre-parsed message type (Msg) for
// subscriptions. It enables efficient, type-safe communications at
// the expense of greater initial complexity and stricter contract
// enforcement.
type AgentMsg interface {
	Subscribe(prefix []Msg) (subId []byte, msgbuf chan Msg, err error)
	Unsubscribe(subId []byte) (err error)
}
