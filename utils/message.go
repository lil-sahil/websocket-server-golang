package utils

import (
	"errors"
	"fmt"
	"net"
)

type RecieveMessage struct {
	c net.Conn
	m []byte
}

func NewRecieveMessage(c net.Conn) *RecieveMessage {
	return &RecieveMessage{
		c: c,
	}
}

func (rm *RecieveMessage) HandleReciveMessage() error {
	rm.m = make([]byte, 20)

	_, err := rm.c.Read(rm.m)

	if err != nil {
		fmt.Println(err)
		rm.c.Close()
	}

	fmt.Println(rm.m)

	// handle FIN bit
	if !rm.getFinBit() {
		// TODO: handle the case if the message is not final
		return errors.New("not yet implemented handling of streaming messages")
	}

	// // handle RSV bits
	// if !rm.handleRSVBits() {
	// 	// TODO: handle the case if the rsv bits are not 0
	// 	return errors.New("not yet implemented handling of non 0 rsv bits")

	// }

	// Handle opcode
	if !rm.handleOpcodeBits() {
		return errors.New("not yet impelemnted handling non text data")
	}

	return nil

}

// Get the fin bit from the recieved message
// if true then fin bit = 1.
func (rm *RecieveMessage) getFinBit() bool {

	// Do a bit wise operation
	// First Byte: 1 0 0 0 0 0 0 1    (129 in decimal)
	// MASK:       1 0 0 0 0 0 0 0    (128 in decimal, 0x80 in hex)
	//             ---------------
	// Result      1 0 0 0 0 0 0 0    (128 in decimal - not zero, so first bit was 1!)
	if rm.m[0]&0x80 != 0 {
		return true
	}

	return false
}

// Handle the RSV bits from the recieved message
func (rm *RecieveMessage) handleRSVBits() bool {
	//Bits 1-3 need to be 0
	if rm.m[0]&0x40 == 0 && rm.m[0]&0x20 == 0 && rm.m[0]&0x10 == 0 {
		return true
	}

	return false
}

// Handle the opcode from the recieved message
func (rm *RecieveMessage) handleOpcodeBits() bool {
	//Bits 4-7 need to be 0001 for text messages.
	// The zeros block the rest, except for the last 4 bits.
	// Byte: 	1 0 0 0 0 0 0 1
	// Mask: 	0 0 0 0 1 1 1 1  (0x0F)
	//       	---------------
	// Result:  0 0 0 0 0 0 0 1  (1 - this the opcode)
	if rm.m[0]&0x0F == 1 {
		return true
	}

	return false
}
