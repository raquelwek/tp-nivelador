package protocol

import (
	"encoding/binary"
	"fmt"
	"strings"

	lottery "github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
)

// / mantains shared logic for BetsPayload and WinnersPayload
type betRecordList struct {
	Records   []lottery.Bet
	batchSize int
}

func (b *betRecordList) AddBet(bet lottery.Bet) error {
	if len(b.Records) >= b.batchSize {
		return fmt.Errorf("cannot add bet: batch is full (max %d)", b.batchSize)
	}
	b.Records = append(b.Records, bet)
	return nil
}

func (b *betRecordList) MarshalPayload() ([]byte, error) {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(len(b.Records))) //@TO DO; check if its useful xd

	for _, bet := range b.Records {
		record := marshalBetRecord(bet)
		buf = append(buf, record...)
	}
	return buf, nil
}

func (p *betRecordList) UnmarshalPayload(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("payload too short for bet count")
	}
	count := binary.BigEndian.Uint16(data[0:2])
	offset := 2

	for i := 0; i < int(count); i++ {
		bet, new_offset := UnmarshalBetRecord(offset, data)
		p.Records = append(p.Records, bet)
		offset = new_offset
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
func marshalBetRecord(bet lottery.Bet) []byte {
	payload := make([]byte, 0)

	documentBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(documentBytes, uint32(bet.Document))
	payload = append(payload, documentBytes...)

	numberBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(numberBytes, uint16(bet.Number))
	payload = append(payload, numberBytes...)

	nameBytes := []byte(bet.FirstName)
	lastNameBytes := []byte(bet.LastName)

	payload = append(payload, byte(len(nameBytes)))
	payload = append(payload, byte(len(lastNameBytes)))

	birthdateBytes := []byte(bet.Birthdate)
	birthdateBytes = append(
		birthdateBytes,
		make([]byte, 8-len(birthdateBytes))...,
	)
	payload = append(payload, birthdateBytes...)

	payload = append(payload, nameBytes...)
	payload = append(payload, lastNameBytes...)

	return payload
}

func UnmarshalBetRecord(offset int, payload []byte) (lottery.Bet, int) {
	document := int(binary.BigEndian.Uint32(payload[offset : offset+4]))
	offset += 4

	number := int(binary.BigEndian.Uint16(payload[offset : offset+2]))
	offset += 2

	nameLength := int(payload[offset])
	offset += 1

	lastNameLength := int(payload[offset])
	offset += 1

	birthdate := strings.TrimRight(
		string(payload[offset:offset+8]),
		"\x00",
	)
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
