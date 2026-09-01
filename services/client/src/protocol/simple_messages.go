package protocol

type ErrorPayload struct {
	Message string
}

func CreateErrorPayload(message string) *ErrorPayload {
	return &ErrorPayload{
		Message: message,
	}
}

func (p *ErrorPayload) MarshalPayload() ([]byte, error) {
	return []byte(p.Message), nil
}

func (p *ErrorPayload) UnmarshalPayload(data []byte) error {
	p.Message = string(data)
	return nil
}

func (p *ErrorPayload) Type() MessageType {
	return ERROR
}

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
