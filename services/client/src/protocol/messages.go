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

const HeaderLength = 6

// String retorna el nombre descriptivo del tipo de mensaje
func (mt MessageType) String() string {
	switch mt {
	case BETS:
		return "BETS"
	case ALL_SENDED:
		return "ALL_SENDED"
	case WINNERS:
		return "WINNERS"
	case ERROR:
		return "ERROR"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", mt)
	}
}

type Payload interface {
	MarshalPayload() ([]byte, error)
	UnmarshalPayload([]byte) error
	Type() MessageType
}

type Message interface {
	Marshal() ([]byte, error)
	GetPayload() Payload
}
type MessageImpl struct {
	AgencyID byte
	Payload  Payload
}

func CreateMessage(agencyID byte, payload Payload) Message {
	return &MessageImpl{
		AgencyID: agencyID,
		Payload:  payload,
	}
}
func (m *MessageImpl) Marshal() ([]byte, error) {
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

func (m *MessageImpl) GetPayload() Payload {
	return m.Payload
}

// @ To do revisar batchsize
func UnmarshalMessage(data []byte, batchSize int) (Message, error) {
	if len(data) < HeaderLength {
		return nil, fmt.Errorf("data too short to be a valid message header")
	}

	msgType := MessageType(data[0])
	agencyID := data[1]
	payloadLength := binary.BigEndian.Uint32(data[2:6])

	if uint32(len(data)) < uint32(HeaderLength)+payloadLength {
		return nil, fmt.Errorf("data too short for declared payload length")
	}

	var payload Payload
	switch msgType {
	case BETS:
		payload = CreateBetsPayload(batchSize)
	case ALL_SENDED:
		payload = CreateAllSendedPayload()
	case WINNERS:
		payload = CreateWinnersPayload(batchSize)
	case ERROR:
		payload = CreateErrorPayload("")
	default:
		return nil, fmt.Errorf("unknown message type: %v", msgType)
	}

	rawPayload := data[HeaderLength : HeaderLength+payloadLength]
	if err := payload.UnmarshalPayload(rawPayload); err != nil {
		return nil, err
	}

	return &MessageImpl{AgencyID: agencyID, Payload: payload}, nil
}
