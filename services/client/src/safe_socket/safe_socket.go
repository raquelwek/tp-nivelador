package safe_socket

import "io"

func SendAll(socket io.Writer, bytes []byte) error {
	bytesSent := 0
	for bytesSent < len(bytes) {
		n, err := socket.Write(bytes[bytesSent:])
		if err != nil {
			return err
		}
		bytesSent += n
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	_, err := io.ReadFull(socket, buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}
