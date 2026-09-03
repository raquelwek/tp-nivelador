package client

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/model"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 10
const CONNECTION_ATTEMPS_DELAY_MS = 200

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   int
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

func (client *Client) Run(ctx context.Context) error {
	const mainAction = "test-lottery-server"
	go sigtermHandler(ctx, client)
	agency := model.CreateAgency(client.config.AgencyId, client.config.OutputFile, client.config.InputFile)
	err := client.processBets(agency, ctx)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		logger.Error(mainAction, logger.Fail, "could not process bets", err)
		return err
	}
	allSended := protocol.CreateMessage(byte(agency.GetId()), protocol.CreateAllSendedPayload())
	sendErr := client.send(allSended)

	if sendErr != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		logger.Error(mainAction, logger.Fail, "could not send all sended message", sendErr)
		return sendErr

	}

	winners, err := client.receive()
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		logger.Error(mainAction, logger.Fail, "could not receive winners message", err)
		return err
	}
	if winners.GetPayload().Type() != protocol.WINNERS {
		return errors.New("invalid payload type")
	}

	err = agency.StoreWinner(winners.GetPayload().(*protocol.WinnersPayload).GetWinners())
	if err != nil {
		logger.Error(mainAction, logger.Fail, "could not store winners", err)
		return err
	}

	logger.Info(mainAction, logger.Success)
	return nil
}

func (client *Client) processBets(agency model.Agency, ctx context.Context) error {
	mainAction := "process-bets"
	messageBetsAmount := 0
	payload := protocol.CreateBetsPayload(client.config.BatchSize)

	for bet, err := range agency.LoadBets() {
		select {
		case <-ctx.Done():
			logger.Info("process-bets", "SIGTERM received, stopping bet processing")
			return ctx.Err()
		default:
		}
		if err != nil {
			return err
		}

		if err := payload.AddBet(bet); err != nil {
			if err := client.sendBetsAndReceiveAck(agency, payload, &messageBetsAmount); err != nil {
				return err
			}
			payload = protocol.CreateBetsPayload(client.config.BatchSize)
			if err := payload.AddBet(bet); err != nil {
				return err
			}
		}
	}
	if len(payload.Records) > 0 {
		if err := client.sendBetsAndReceiveAck(agency, payload, &messageBetsAmount); err != nil {
			return err
		}
	}

	logger.Info(mainAction, logger.Success, "messages-bets-sent", messageBetsAmount)
	return nil
}

// Marshall the message and send it through the socket. Returns an error if the message could not be sent.
func (client *Client) send(message protocol.Message) error {
	mainAction := "send-message"
	messageArgs := []any{"message-type", message.GetPayload().Type().String()}
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

// Receives a message with the structure indicated in the protocol
// Returns the message and an error if the message could not be received or unmarshalled.
func (client *Client) receive() (protocol.Message, error) {
	mainAction := "receive-message"
	logger.Info(mainAction, logger.InProgress)

	bytes_header, err := safe_socket.RecvAll(client.conn, protocol.HeaderLength)
	if err != nil {
		logger.Error("receive-message", logger.Fail, "err", err)
		return nil, err
	}
	payloadLength := (int)(binary.BigEndian.Uint32(bytes_header[2:6]))
	bytes_payload, err := safe_socket.RecvAll(client.conn, payloadLength)
	if err != nil {
		logger.Error("receive-message", logger.Fail, "err", err)
		return nil, err
	}
	message, err := protocol.UnmarshalMessage(append(bytes_header, bytes_payload...), client.config.BatchSize)
	if err != nil {
		logger.Error("unmarshal-message", logger.Fail, "err", err)
		return nil, err
	}
	messageArgs := []any{"message-type", message.GetPayload().Type().String(), "payload-length", payloadLength}
	logger.Info(mainAction, logger.Success, messageArgs...)
	return message, nil

}

func (client *Client) sendBetsAndReceiveAck(agency model.Agency, payload *protocol.BetsPayload, messageBetsAmount *int) error {
	message := protocol.CreateMessage(byte(agency.GetId()), payload)
	err := client.send(message)
	if err != nil {
		return err
	}
	*messageBetsAmount++

	return client.receiveAck()
}
func (client *Client) receiveAck() error {
	msg, err := client.receive()
	if err != nil {
		logger.Error("receive-ack", logger.Fail, "err", err)
		return err
	}
	if msg != nil && msg.GetPayload().Type() != protocol.ACK {
		logger.Error("receive-ack", logger.Fail, "err", "invalid payload type")
		return errors.New("unexpected payload type, expected ACK")
	}
	return nil
}
func (client *Client) Close() {
	const mainAction = "close-client"
	logger.Info(mainAction, logger.InProgress)
	if client.conn != nil {
		client.conn.Close()
	}
	logger.Info(mainAction, logger.Success)
}

func sigtermHandler(ctx context.Context, client *Client) {
	logger.Info("sigterm-received", logger.InProgress)
	<-ctx.Done()
	client.Close()
}
