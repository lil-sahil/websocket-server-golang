package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

type RecieveMessage struct {
	c      net.Conn
	m      []byte
	buffer bytes.Buffer // buffer to store the bytes as they are recieved. This is used in the event when the FIN bit is not 1.
}

func NewRecieveMessage(c net.Conn) *RecieveMessage {
	return &RecieveMessage{
		c: c,
	}
}

func (rm *RecieveMessage) HandleReciveMessage() error {
	// Read the first two bytes to get FIN, RSV1,2,3 and OPCODE values
	rm.m = make([]byte, 2)

	_, err := rm.c.Read(rm.m)

	if err != nil {
		fmt.Println(err)
		rm.c.Close()
	}

	fmt.Println(rm.m)

	// handle FIN bit
	if !rm.getFinBit() {
		// handle the case if the message is not final
		rm.buffer.Write(rm.m)

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

	// ---
	// Logic to check payload length
	// ---

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

// Determine MASK
// func (rm *RecieveMessage) determineMask() {
// 	if rm.m[1]&0x80 != 0 {
// 		// Therefore it is masked
// 	}
// }

// Decode the payload length
func (rm *RecieveMessage) decodePayload() (uint64, error) {
	payloadLen := rm.m[1] & 0x7f

	if payloadLen <= 125 {
		return uint64(payloadLen), nil
	}

	if payloadLen == 126 {
		// Read the next 16 bits
		b := make([]byte, 2)

		_, err := rm.c.Read(b)

		if err != nil {
			return 0, err
		}

		length := binary.BigEndian.Uint16(b)

		return uint64(length), nil
	}

	// Read the next 64 bits = 8 bytes
	b := make([]byte, 8)
	_, err := rm.c.Read(b)

	if err != nil {
		return 0, err
	}

	// Check to make sure that the most significant bit is 0
	if b[0]&0x80 != 0 {
		return 0, errors.New("the most significant bit is not 0")
	}

	length := binary.BigEndian.Uint64(b)

	return length, nil
}
