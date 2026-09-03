package protocol

type FullBatchError struct{}

func (e *FullBatchError) Error() string {
	return "cannot add bet: batch is full"
}

type InvalidPayloadError struct {
	mesage string
}

func (e *InvalidPayloadError) Error() string {
	return "invalid payload: " + e.mesage
}
