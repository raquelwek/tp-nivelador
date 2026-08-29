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

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

const BATCH_SIZE = 512
const ECHO_CLIENT_MESSAGE_AMOUNT = 3
const ECHO_CLIENT_MESSAGE_DELAY_MS = 1000

const FILE_PERMISSIONS_CODE = 0644

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
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

	agencyBets, err := agency.GetBets()
	if err != nil {
		logger.Error(mainAction, logger.Fail, "err", err)
		return err
	}
	payload := protocol.CreateBetsPayload(BATCH_SIZE)
	for _, bet := range agencyBets {
		err := payload.AddBet(bet)
		if err.Error() == "cannot add bet: batch is full" {
			message := protocol.CreateMessage(byte(agencyId), payload)
			client.send(message, mainAction)
			err = client.send(message, mainAction)
			if err != nil {
				logger.Error(mainAction, logger.Fail, "err", err)
				return err
			}
			payload = protocol.CreateBetsPayload(BATCH_SIZE)
		}
	}
	allSended := protocol.CreateMessage(byte(agencyId), protocol.CreateAllSendedPayload())
	sendErr := client.send(allSended, mainAction)
	if sendErr != nil {
		logger.Error(mainAction, logger.Fail, "err", sendErr)
		return sendErr
	}

	winners, err := client.receive(mainAction)
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

func (client *Client) send(message protocol.Message, mainAction string) error {
	messageArgs := []any{"agency-id", client.config.AgencyId}
	logger.Info(mainAction, logger.InProgress, messageArgs...)

	bytes, err := message.Marshal()
	if err != nil {
		logger.Error("marshal-message", logger.Fail, messageArgs...)
		return err
	}

	if err := safe_socket.SendAll(client.conn, bytes); err != nil {
		logger.Error("send-message", logger.Fail, messageArgs...)
		return err
	}

	return nil
}

func (client *Client) receive(mainAction string) (protocol.Message, error) {
	messageArgs := []any{"agency-id", client.config.AgencyId}
	logger.Info(mainAction, logger.InProgress, messageArgs...)

	bytes_header, err := safe_socket.RecvAll(client.conn, protocol.HeaderLength)
	payloadLength := (int)(binary.BigEndian.Uint32(bytes_header[2:6]))
	bytes_payload, err := safe_socket.RecvAll(client.conn, payloadLength)
	if err != nil {
		logger.Error("receive-message", logger.Fail, messageArgs...)
		return nil, err
	}
	message, err := protocol.UnmarshalMessage(append(bytes_header, bytes_payload...), BATCH_SIZE)
	if err != nil {
		logger.Error("unmarshal-message", logger.Fail, messageArgs...)
		return nil, err
	}

	logger.Info(mainAction, logger.Success, messageArgs...)
	return message, nil

}
