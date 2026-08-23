package client

import (
	"bufio"
	"net"
	"os"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

const ECHO_CLIENT_BUFFER_SIZE = 512
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
	const mainAction = "test-echo-server"
	defer client.conn.Close()

	file, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("open-file", logger.Fail, "err", err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		err := client.sendBet(line, mainAction)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)
	return nil
}

func (client *Client) sendBet(line string, mainAction string) error {
	messageArgs := []any{"agency-id", client.config.AgencyId}
	logger.Info(mainAction, logger.InProgress, messageArgs...)

	if err := safe_socket.SendAll(client.conn, []byte(line)); err != nil {
		logger.Error("send-message", logger.Fail, messageArgs...)
		return err
	}

	responseBuffer, err := safe_socket.RecvAll(client.conn, ECHO_CLIENT_BUFFER_SIZE)
	if err != nil {
		logger.Error("recv-response", logger.Fail, messageArgs...)
		return err
	}

	if err := client.persistResponse(responseBuffer); err != nil {
		return err
	}

	if string(responseBuffer) != line {
		logger.Error("check-response", logger.Fail, messageArgs...)
		return err
	}

	time.Sleep(ECHO_CLIENT_MESSAGE_DELAY_MS * time.Millisecond)

	return nil
}

func (client *Client) persistResponse(content []byte) error {
	const action = "persist-response"

	if err := os.WriteFile(client.config.OutputFile, content, FILE_PERMISSIONS_CODE); err != nil {
		logger.Error(action, logger.Fail, "err", err)
		return err
	}
	logger.Info(action, logger.Success)
	return nil
}
