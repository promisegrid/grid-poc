package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/fxamacker/cbor/v2"
	. "github.com/stevegt/goadapt"
)

func main() {
	wrapped := wrap()
	show(wrapped, "wrapped")

	labeled := label()
	show(labeled, "labeled")
}

func show(msg []byte, fn string) {
	Pf("%s:\n", fn)
	// show the message data in hex
	for _, b := range msg {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()

	// show the message data in ASCII
	for _, b := range msg {
		// if not in ASCII range, show a dot
		if b < 32 || b > 126 {
			b = '.'
		}
		fmt.Printf(" %c ", b)
	}
	fmt.Println()

	// show the first byte in binary
	fmt.Printf("%08b\n", msg[0])
	// Output: 01101011

	// show the CBOR major type
	fmt.Printf("%d\n", msg[0]>>5)
	// Output: 3

	// show diagnostic notation for all CBOR data items in msg (might
	// be a sequence of data items)
	tmpMsg := msg[:]
	for i := 0; len(tmpMsg) > 0; i++ {
		var notation string
		var err error
		notation, tmpMsg, err = cbor.DiagnoseFirst(tmpMsg)
		Ck(err)
		fmt.Printf("%d: %s\n", i, notation)
	}

	// write the message to a file
	path := fmt.Sprintf("/tmp/%s.cbor", fn)
	err := ioutil.WriteFile(path, msg, 0644)
	Ck(err)

	Pf("saved to %s\n", path)
	Pl()
}

func wrap() []byte {
	// protocol name "GRID"
	// hex 47 52 49 44
	// decimal 1196575044

	// Define the protocol number and data
	var protocolNumber uint64 = 1196575044

	payload := "hello world"

	// wrap the payload in the protocol number tag
	protoTagged := cbor.Tag{Number: protocolNumber, Content: payload}

	// wrap the tagged payload in a self-describing CBOR message
	cborTagged := cbor.Tag{Number: 55799, Content: protoTagged}

	// Create an encoding mode with custom tag support
	opts := cbor.EncOptions{}
	em, err := opts.EncMode()
	Ck(err)

	// Marshal the tagged data to CBOR
	msg, err := em.Marshal(cborTagged)
	Ck(err)

	return msg
}

func label() []byte {
	// protocol name "GRID"
	// hex 47 52 49 44
	// decimal 1196575044

	// Define the protocol number and data
	var protocolNumber uint64 = 1196575044

	// create the sequence label for the protocol number
	protoTag := cbor.Tag{Number: protocolNumber, Content: "BOR"}
	cborTag := cbor.Tag{Number: 55800, Content: protoTag}

	payload := "hello world"

	// Create an encoding mode with custom tag support
	opts := cbor.EncOptions{}
	em, err := opts.EncMode()
	Ck(err)

	buf := bytes.Buffer{}

	// send the sequence label
	msg, err := em.Marshal(cborTag)
	Ck(err)
	_, err = buf.Write(msg)
	Ck(err)

	// send the payload
	msg, err = em.Marshal(payload)
	Ck(err)
	_, err = buf.Write(msg)
	Ck(err)

	return buf.Bytes()
}
