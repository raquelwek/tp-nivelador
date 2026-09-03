package client

import (
	"encoding/binary"
	"net"
	"strconv"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 10
const CONNECTION_ATTEMPS_DELAY_MS = 200

const ECHO_CLIENT_MESSAGE_AMOUNT = 3
const ECHO_CLIENT_MESSAGE_DELAY_MS = 1000

const FILE_PERMISSIONS_CODE = 0644

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run() error {
	const mainAction = "test-lottery-server"
	agencyId, err := strconv.Atoi(client.config.AgencyId)
	if err != nil {
		logger.Error(mainAction, logger.Fail, "err", err)
		return err
	}
	agency := model.CreateAgency(agencyId, client.config.OutputFile, client.config.InputFile)

	messageBetsAmount := 0
	payload := protocol.CreateBetsPayload(client.config.BatchSize)
	for bet, err := range agency.LoadBets() {
		if err != nil {
			logger.Error(mainAction, logger.Fail, "err", err)
			return err
		}

		err1 := payload.AddBet(bet)
		logger.Info(mainAction, logger.InProgress, "agency-id", client.config.AgencyId, "bet added", bet)
		if err1 != nil && err1.Error() == "cannot add bet: batch is full" {

			// Envía el payload lleno
			message := protocol.CreateMessage(byte(agencyId), payload)
			err = client.send(message)
			if err != nil {
				logger.Error(mainAction, logger.Fail, "err", err)
				return err
			}
			client.receiveAck()
			messageBetsAmount++
			// Crea nuevo payload y agrega la apuesta que causó el error
			payload = protocol.CreateBetsPayload(client.config.BatchSize)
			err1 = payload.AddBet(bet)
			if err1 != nil {
				logger.Error(mainAction, logger.Fail, "err", err1)
				return err1
			}
		}
	}
	// Envía lo que quedó en el payload
	message := protocol.CreateMessage(byte(agencyId), payload)
	err = client.send(message)
	if err != nil {
		logger.Error(mainAction, logger.Fail, "err", err)
		return err
	}
	messageBetsAmount++
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId, "messages-bets-sent", messageBetsAmount)
	allSended := protocol.CreateMessage(byte(agencyId), protocol.CreateAllSendedPayload())
	client.receiveAck()
	sendErr := client.send(allSended)
	if sendErr != nil {
		logger.Error(mainAction, logger.Fail, "err", sendErr)
		return sendErr
	}

	winners, err := client.receive()
	if err != nil {
		return err
	}
	winnersPayload, ok := winners.GetPayload().(*protocol.WinnersPayload)
	if !ok {
		logger.Error(mainAction, logger.Fail, "err", "invalid payload type")
		return err
	}

	err = agency.StoreWinner(winnersPayload.GetWinners())
	if err != nil {
		logger.Error(mainAction, logger.Fail, "err", err)
		return err
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)
	return nil
}

func (client *Client) send(message protocol.Message) error {
	mainAction := "send-message"
	messageArgs := []any{"agency-id", client.config.AgencyId, "message-type", message.GetPayload().Type().String()}
	logger.Info(mainAction, logger.InProgress, messageArgs...)

	bytes, err := message.Marshal()
	if err != nil {
		logger.Error("marshal-message", logger.Fail, messageArgs...)
		return err
	}

	logger.Info(mainAction, logger.InProgress, append(messageArgs, "bytes-count", len(bytes))...)
	if err := safe_socket.SendAll(client.conn, bytes); err != nil {
		logger.Error("send-message", logger.Fail, messageArgs...)
		return err
	}
	logger.Info(mainAction, logger.Success, messageArgs...)
	return nil
}
func (client *Client) receiveAck() error {
	msg, err := client.receive()
	if err != nil {
		logger.Error("receive-ack", logger.Fail, "err", err)
		return err
	}
	if msg != nil && msg.GetPayload().Type() != protocol.ACK {
		logger.Error("receive-ack", logger.Fail, "err", "invalid payload type")
		return nil
	}
	return nil

}
func (client *Client) receive() (protocol.Message, error) {
	mainAction := "receive-message"
	messageArgs := []any{"agency-id", client.config.AgencyId}
	logger.Info(mainAction, logger.InProgress, messageArgs...)

	bytes_header, err := safe_socket.RecvAll(client.conn, protocol.HeaderLength)
	if err != nil {
		logger.Error("receive-message", logger.Fail, messageArgs...)
		return nil, err
	}
	payloadLength := (int)(binary.BigEndian.Uint32(bytes_header[2:6]))
	bytes_payload, err := safe_socket.RecvAll(client.conn, payloadLength)
	if err != nil {
		logger.Error("receive-message", logger.Fail, messageArgs...)
		return nil, err
	}
	message, err := protocol.UnmarshalMessage(append(bytes_header, bytes_payload...), client.config.BatchSize)
	if err != nil {
		logger.Error("unmarshal-message", logger.Fail, messageArgs...)
		return nil, err
	}
	messageArgs = append(messageArgs, "message-type", message.GetPayload().Type().String(), "payload-length", payloadLength)
	logger.Info(mainAction, logger.Success, messageArgs...)
	return message, nil

}
