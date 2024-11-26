package utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/lil-sahil/websocket-server-golang/types"
)

type frame struct {
	header        []byte
	fin           bool
	rsv1          bool
	rsv2          bool
	rsv3          bool
	opccode       byte
	payloadLength uint64
	payloadData   []byte
	maskingKey    []byte
}

type RecieveMessage struct {
	c     net.Conn
	frame frame
	cbs   map[types.CallbackEvent]func(string)
}

type SendMessage struct {
	payloadData string
}

func NewRecieveMessage(c net.Conn, cbs map[types.CallbackEvent]func(string)) *RecieveMessage {
	return &RecieveMessage{
		c:   c,
		cbs: cbs,
	}
}

func NewSendMessage(payloadData string) *SendMessage {
	return &SendMessage{
		payloadData: payloadData,
	}
}

func (sm *SendMessage) CreateFrame() []byte {
	// Determine the length of the message
	mLen := len(sm.payloadData)
	payLoad := []byte(sm.payloadData)

	// Set the fin bit - assume that it is 1 for now  and rsv bits as 000. = 1000 this is 0x8
	// Set the opcode as 0001 this is 0x1
	// Therefore the first byte of the frame will be 1000 0001 = 0x81
	b := []byte{0x81}

	// if the payload length is less than or equal to 125, the payload length is stored in the 7bits of the 2nd byte.
	// The first bit of the 2nd byte is used for masking.
	// The reason for this is because 7**2 = 128 which means that it can stored in the 7 availaible bits of the 2nd byte.

	// If the payload length is larger than 125 but less than 65535, then the 2nd byte is set to 126.
	// This is done to let the client/server that is decoding the message frame to know that the exact payload length is found in the next two bytes or 16 bits.
	// 2**16 = 65535

	// If the payload is larger than 65535, then the 2nd byte is set to 127.
	// This is done to let the client/server that is decoding the message frame to konw that the exact payload length is found in the next 8 bytes or 64 bits.

	switch {
	case mLen <= 125:
		b = append(b, byte(mLen))
	case mLen <= 65535:
		// Set the 2nd byte as 126.
		// The next 2 bytes need to be set as the payload length
		// In order to do this, payloadLength >> 8 will get the Most Significant Byte. 0xFF will get the Least Significant Byte
		// ex: Suppose we want to encode the value 260 into two bytes - which we know to be 00000001 00000100
		// 260 >> 8 = 00000001
		// 260 & 0xFF = 00000001 00000100
		//                       11111111
		//            =          00000100
		b = append(b, 126, byte(mLen>>8), byte(mLen&0xFF))
	default:
		b = append(b, 127, byte(mLen>>54), byte(mLen>>48), byte(mLen>>40), byte(mLen>>32), byte(mLen>>24), byte(mLen>>16), byte(mLen>>8), byte(mLen&0xFF))
	}

	b = append(b, payLoad...)

	return b
}

func (sm *SendMessage) SendMessage(c net.Conn) {
	frame := sm.CreateFrame()
	fmt.Println(frame)
	c.Write(frame)
}

func (rm *RecieveMessage) HandleReciveMessage() error {
	// Read the first two bytes to get FIN, RSV1,2,3 and OPCODE values
	rm.frame.header = make([]byte, 2)

	_, err := rm.c.Read(rm.frame.header)

	if err != nil {
		fmt.Println(err)
		rm.c.Close()
	}

	// Get FIN bit
	rm.getFinBit()

	if !rm.frame.fin {
		// handle the case if the message is not final
		return errors.New("not yet implemented handling of streaming messages")
	}

	// Get the RSV Bits
	rm.handleRSVBits()

	if rm.frame.rsv1 == false || rm.frame.rsv2 == false || rm.frame.rsv3 == false {
		// TODO: handle the case if the rsv bits are not 0
		return errors.New("not yet implemented handling of non 0 rsv bits")

	}

	// Get opcode bits
	rm.handleOpcodeBits()

	// To-do: handle opcode bit

	// Determine if the message is masked
	if !rm.determineMask() {
		// It is not masked therefore return error
		return errors.New("message is not masked")
	}

	// Logic to check payload length
	err = rm.calculatePayloadLength()

	if err != nil {
		return err
	}

	// Decode message
	rm.decodeMessage()

	// Fire off callbacks
	for key, cb := range rm.cbs {
		if key == "message" {
			cb(string(rm.frame.payloadData))
		}
	}

	return nil

}

// Get the fin bit from the recieved message
// if true then fin bit = 1.
func (rm *RecieveMessage) getFinBit() {

	// Do a bit wise operation
	// First Byte: 1 0 0 0 0 0 0 1    (129 in decimal)
	// MASK:       1 0 0 0 0 0 0 0    (128 in decimal, 0x80 in hex)
	//             ---------------
	// Result      1 0 0 0 0 0 0 0    (128 in decimal - not zero, so first bit was 1!)
	if rm.frame.header[0]&0x80 != 0 {
		rm.frame.fin = true
		return
	}

	rm.frame.fin = false
}

// Handle the RSV bits from the recieved message
func (rm *RecieveMessage) handleRSVBits() {
	//Bits 1-3 need to be 0
	if rm.frame.header[0]&0x40 == 0 && rm.frame.header[0]&0x20 == 0 && rm.frame.header[0]&0x10 == 0 {
		rm.frame.rsv1 = true
		rm.frame.rsv2 = true
		rm.frame.rsv3 = true
		return
	}

	rm.frame.rsv1 = false
	rm.frame.rsv2 = false
	rm.frame.rsv3 = false
}

// Handle the opcode from the recieved message
func (rm *RecieveMessage) handleOpcodeBits() {
	//Bits 4-7 need to be 0001 for text messages.
	// The zeros block the rest, except for the last 4 bits.
	// Byte: 	1 0 0 0 0 0 0 1
	// Mask: 	0 0 0 0 1 1 1 1  (0x0F)
	//       	---------------
	// Result:  0 0 0 0 0 0 0 1  (1 - this the opcode)
	rm.frame.opccode = rm.frame.header[0] & 0x0F
}

// Determine MASK
func (rm *RecieveMessage) determineMask() bool {
	if rm.frame.header[1]&0x80 != 0 {
		// Therefore it is masked
		return true
	}

	return false
}

// Decode the message
func (rm *RecieveMessage) decodeMessage() {
	maskingKey := make([]byte, 4)
	rm.c.Read(maskingKey)

	rm.frame.maskingKey = maskingKey

	rm.frame.payloadData = make([]byte, rm.frame.payloadLength)

	rm.c.Read(rm.frame.payloadData)
	// Decode payload data
	for i, octet := range rm.frame.payloadData {
		rm.frame.payloadData[i] = octet ^ rm.frame.maskingKey[i%4]
	}

	fmt.Println(string(rm.frame.payloadData))

}

// get the payload length
func (rm *RecieveMessage) calculatePayloadLength() error {
	payloadLen := rm.frame.header[1] & 0x7f

	if payloadLen <= 125 {
		rm.frame.payloadLength = uint64(payloadLen)
		return nil
	}

	if payloadLen == 126 {
		// Read the next 16 bits and store in internal buffer
		b := make([]byte, 2)

		_, err := rm.c.Read(b)

		if err != nil {
			return err
		}
		rm.frame.payloadLength = uint64(binary.BigEndian.Uint16(b))
		return nil
	}

	// Read the next 64 bits = 8 bytes
	b := make([]byte, 8)
	_, err := rm.c.Read(b)

	if err != nil {
		return err
	}

	// Check to make sure that the most significant bit is 0
	if b[0]&0x80 != 0 {
		return errors.New("the most significant bit is not 0")
	}

	rm.frame.payloadLength = binary.BigEndian.Uint64(b)

	return nil
}
