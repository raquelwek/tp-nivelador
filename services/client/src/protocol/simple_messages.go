package protocol

type AllSendedPayload struct {
}

func CreateAllSendedPayload() *AllSendedPayload {
	return &AllSendedPayload{}
}
func (p *AllSendedPayload) MarshalPayload() ([]byte, error) {
	return []byte{}, nil
}

func (p *AllSendedPayload) UnmarshalPayload(data []byte) error {
	return nil
}

func (p *AllSendedPayload) Type() MessageType {
	return ALL_SENDED
}

type AckPayload struct{}

func CreateAckPayload() *AckPayload {
	return &AckPayload{}
}

func (p *AckPayload) MarshalPayload() ([]byte, error) {
	return []byte{}, nil
}

func (p *AckPayload) UnmarshalPayload(data []byte) error {
	return nil
}

func (p *AckPayload) Type() MessageType {
	return ACK
}
