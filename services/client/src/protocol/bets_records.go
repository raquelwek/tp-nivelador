package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"

	lottery "github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
)

// / mantains shared logic for BetsPayload and WinnersPayload
type betRecordList struct {
	Records   []lottery.Bet
	batchSize int
}

const (
	documentLen  = 4
	numberLen    = 2
	fnLenLen     = 1
	lnLenLen     = 1
	birthdateLen = 8
	headerLen    = documentLen + numberLen + fnLenLen + lnLenLen + birthdateLen // 16
)

// Adds a bet to the list. Returns an error if the batch is full.
func (b *betRecordList) AddBet(bet lottery.Bet) error {
	if len(b.Records) >= b.batchSize {
		return &FullBatchError{}
	}
	b.Records = append(b.Records, bet)
	return nil
}

func (b *betRecordList) MarshalPayload() ([]byte, error) {
	// Pre-calculate total size to avoid per-bet allocations
	total := 2 // 2-byte count header
	for _, bet := range b.Records {
		total += headerLen + len(bet.FirstName) + len(bet.LastName)
	}

	buf := make([]byte, total)
	binary.BigEndian.PutUint16(buf[0:2], uint16(len(b.Records)))
	offset := 2

	for _, bet := range b.Records {
		offset = marshalBetRecordInto(buf, offset, bet)
	}
	return buf, nil
}

func (p *betRecordList) UnmarshalPayload(data []byte) error {
	if len(data) < 2 {
		return &InvalidPayloadError{mesage: "payload too short"}
	}
	count := binary.BigEndian.Uint16(data[0:2])
	offset := 2

	for i := 0; i < int(count); i++ {
		bet, new_offset := UnmarshalBetRecord(offset, data)
		p.Records = append(p.Records, bet)
		offset = new_offset
	}
	if len(p.Records) != int(count) {
		return &InvalidPayloadError{mesage: fmt.Sprintf("bets count mismatch: expected %d, got %d", count, len(p.Records))}
	}
	return nil
}

type BetsPayload struct {
	betRecordList
}

func CreateBetsPayload(batchSize int) *BetsPayload {
	return &BetsPayload{betRecordList{batchSize: batchSize}}
}

func (p *BetsPayload) Type() MessageType { return BETS }

type WinnersPayload struct {
	betRecordList
}

func (p *WinnersPayload) Type() MessageType { return WINNERS }

func CreateWinnersPayload(batchSize int) *WinnersPayload {
	return &WinnersPayload{betRecordList{batchSize: batchSize}}
}
func (p *WinnersPayload) GetWinners() []lottery.Bet {
	return p.Records
}

// marshalBetRecordInto writes a bet record into buf starting at offset and returns the new offset.
func marshalBetRecordInto(buf []byte, offset int, bet lottery.Bet) int {
	lenName := len(bet.FirstName)
	lenLastName := len(bet.LastName)

	binary.BigEndian.PutUint32(buf[offset:], uint32(bet.Document))
	offset += documentLen

	binary.BigEndian.PutUint16(buf[offset:], uint16(bet.Number))
	offset += numberLen

	buf[offset] = byte(lenName)
	offset += fnLenLen

	buf[offset] = byte(lenLastName)
	offset += lnLenLen

	copy(buf[offset:offset+birthdateLen], bet.Birthdate) // padea con 0x00 si falta
	offset += birthdateLen

	copy(buf[offset:], bet.FirstName)
	offset += lenName

	copy(buf[offset:], bet.LastName)
	offset += lenLastName

	return offset
}
func UnmarshalBetRecord(offset int, payload []byte) (lottery.Bet, int) {
	document := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
	offset += 4

	number := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
	offset += 2

	nameLength := int(payload[offset])
	offset++

	lastNameLength := int(payload[offset])
	offset++

	birthdate := string(bytes.TrimRight(payload[offset:offset+8], "\x00"))
	offset += 8

	name := string(payload[offset : offset+nameLength])
	offset += nameLength

	lastName := string(payload[offset : offset+lastNameLength])
	offset += lastNameLength

	bet := lottery.Bet{
		//AgencyId:  agencyID,
		FirstName: name,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}
	return bet, offset
}
