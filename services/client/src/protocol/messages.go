package protocol

import (
	"encoding/binary"
	"fmt"
)

type MessageType byte

const (
	BETS       MessageType = 0x01
	ALL_SENDED MessageType = 0x02
	WINNERS    MessageType = 0x03
	ERROR      MessageType = 0x04
)

const HeaderLength = 6 // 1 (type) + 1 (agency_id) + 4 (payload_length)

// Payload es lo único que cada tipo de mensaje concreto implementa.
// No sabe nada de type/agency_id/length — solo serializa SU contenido.
type Payload interface {
	MarshalPayload() ([]byte, error)
	UnmarshalPayload([]byte) error
	Type() MessageType
}

// Message es el mensaje completo: header + payload.
// No es una interfaz — es un struct concreto, porque el manejo de header
// es siempre el mismo sin importar el tipo de payload.
type Message struct {
	AgencyID byte
	Payload  Payload
}

func (m *Message) Marshal() ([]byte, error) {
	payloadBytes, err := m.Payload.MarshalPayload()
	if err != nil {
		return nil, err
	}

	header := make([]byte, HeaderLength)
	header[0] = byte(m.Payload.Type())
	header[1] = m.AgencyID
	binary.BigEndian.PutUint32(header[2:6], uint32(len(payloadBytes)))

	return append(header, payloadBytes...), nil
}

// payloadFactory: mapea cada MessageType a un constructor de Payload vacío,
// listo para que UnmarshalPayload lo llene. Es tu "switch", pero como tabla de despacho.
var payloadFactory = map[MessageType]func() Payload{
	BETS:       func() Payload { return &BetsPayload{} },
	ALL_SENDED: func() Payload { return &AllSendedPayload{} },
	WINNERS:    func() Payload { return &WinnersPayload{} },
	ERROR:      func() Payload { return &ErrorPayload{} },
}

func UnmarshalMessage(data []byte) (*Message, error) {
	if len(data) < HeaderLength {
		return nil, fmt.Errorf("data too short to be a valid message header")
	}

	msgType := MessageType(data[0])
	agencyID := data[1]
	payloadLength := binary.BigEndian.Uint32(data[2:6])

	if uint32(len(data)) < uint32(HeaderLength)+payloadLength {
		return nil, fmt.Errorf("data too short for declared payload length")
	}

	newPayload, ok := payloadFactory[msgType]
	if !ok {
		return nil, fmt.Errorf("unknown message type: %d", msgType)
	}

	payload := newPayload()
	rawPayload := data[HeaderLength : HeaderLength+payloadLength]
	if err := payload.UnmarshalPayload(rawPayload); err != nil {
		return nil, err
	}

	return &Message{AgencyID: agencyID, Payload: payload}, nil
}
