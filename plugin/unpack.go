package plugin

import (
	"encoding/binary"
	"errors"
	"log"
	"net"

	"github.com/comp500/caddy-sstp/ppp"
)

func decodeHeader(input []byte) (bool, int, error) {
	if len(input) < 4 {
		log.Print()
		return true, 0, errors.New("Packet not long enough")
	}

	majVer := input[0] >> 4
	minVer := input[0] & 0xf
	isControl := input[1] == 1
	length := int(binary.BigEndian.Uint16(input[2:4]))

	if majVer == 1 && minVer == 0 && length > 4 {
		return isControl, (length - 4), nil
	}

	// isControl, lengthToRead, err
	return true, 0, errors.New("Invalid packet")
}

func parseControl(input []byte) sstpControlHeader {
	controlHeader := sstpControlHeader{}
	controlHeader.MessageType = MessageType(binary.BigEndian.Uint16(input[:2]))
	controlHeader.AttributesLength = binary.BigEndian.Uint16(input[2:4])

	attributes := make([]sstpAttribute, int(controlHeader.AttributesLength))
	consumedBytes := 4
	for i := 0; i < len(attributes); i++ {
		attribute := sstpAttribute{}
		// ignore Reserved byte
		attribute.AttributeID = AttributeID(input[consumedBytes+1])
		attribute.Length = binary.BigEndian.Uint16(input[(consumedBytes + 2):(consumedBytes + 4)])
		attribute.Data = input[(consumedBytes + 4):(consumedBytes + int(attribute.Length))]
		consumedBytes += int(attribute.Length)

		attributes[i] = attribute
	}
	controlHeader.Attributes = attributes
	return controlHeader
}

func handleDataPacket(data []byte, conn net.Conn, pppConnection ppp.Connection) {
	//log.Printf("read: %v\n", dataHeader)
	if pppConnection != nil {
		_, err := pppConnection.Write(data)
		handleErr(err)
		//log.Printf("%v bytes written to pppd", n)
	} else {
		log.Fatal("pppd instance not started")
	}
}

func handleControlPacket(controlHeader sstpControlHeader, conn net.Conn, pppConfig ppp.Config, pppConnection ppp.Connection) *ppp.Connection {
	log.Printf("read: %v\n", controlHeader)

	if controlHeader.MessageType == MessageTypeCallConnectRequest {
		sendConnectionAckPacket(conn)
		// TODO: implement Nak?
		// -> if protocols specified by req not supported
		// however there is only PPP currently, so not a problem
		pppConn, err := ppp.NewConnection(pppConfig)
		handleErr(err)
		log.Print("pppd instance created")
		return pppConn
	} else if controlHeader.MessageType == MessageTypeCallDisconnect {
		sendDisconnectAckPacket(conn)
		err := pppConnection.Close()
		handleErr(err)
	} else if controlHeader.MessageType == MessageTypeEchoRequest {
		// TODO: implement hello timer and echo request?
		sendEchoResponsePacket(conn)
	} else if controlHeader.MessageType == MessageTypeCallAbort {
		// TODO: parse error
		log.Fatal("error encountered, connection aborted")
	}
	// TODO: implement connected

	return nil
}
