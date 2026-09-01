package client

import (
	"os"
	"net"
	"time"
	"bufio"
	"errors"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 500

const ECHO_CLIENT_BUFFER_SIZE = 512
const ECHO_CLIENT_MESSAGE_AMOUNT = 3
const ECHO_CLIENT_MESSAGE_DELAY_MS = 1000

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
}

type Client struct {
	conn    net.Conn
	config 	   ClientConfig
	inputFile  *os.File
	outputFile *os.File	
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}
	inputFile, outputFile, err := openFiles(config.InputFile, config.OutputFile)
	if err != nil {
		logger.Error("open-files", logger.Fail, "err", err)
		return nil, err
	}

	client := &Client{conn: conn, config: config, inputFile: inputFile, outputFile: outputFile}
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

func openFiles(inputFile, outputFile string) (*os.File, *os.File, error) {
	InputFile, err := os.Open(inputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail, "err", err)
		return nil, nil, err
	}

	OutputFile, err := os.OpenFile(outputFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logger.Error("open-output-file", logger.Fail, "err", err)
		InputFile.Close()
		return nil, nil, err
	}

	return InputFile, OutputFile, nil
}

func formatMessage(message string) ([]byte, error) {
	messageLen := len(message)
	if messageLen > 255 {
		err := errors.New("integer overflow")
		logger.Error("format-message", logger.Fail, "message too long", err)
		return nil, err
	}

	return append([]byte{uint8(messageLen)}, []byte(message)...), nil
}

func parseMessage(bytes []byte) ([]byte, error) {
	if len(bytes) < 1 {
		return nil, errors.New("empty buffer")
	}
	messageLen := int(bytes[0])
	if len(bytes) < 1+messageLen {
    	return nil, errors.New("response buffer is shorter than message len")
	}
	return append(bytes[1 : 1+messageLen],'\r','\n'), nil
}


func (client *Client) Run() error {
	const mainAction = "test-echo-server"
	defer client.conn.Close()
	defer client.inputFile.Close()
	defer client.outputFile.Close()

	scanner := bufio.NewScanner(client.inputFile)
	writer := bufio.NewWriter(client.outputFile)
	
	messageId := 0
	for scanner.Scan() {
		messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", messageId}
		logger.Info(mainAction, logger.InProgress, messageArgs...)
		
		clientMessage, err := formatMessage(scanner.Text())
		if err != nil {
			logger.Error("format-message", logger.Fail, messageArgs...)
			return err
		}

		if err := safe_socket.SendAll(client.conn, clientMessage); err != nil {
			logger.Error("send-message", logger.Fail, messageArgs...)
			return err
		}

		responseBuffer, err := safe_socket.RecvAll(client.conn, ECHO_CLIENT_BUFFER_SIZE)
		if err != nil {
			logger.Error("recv-response", logger.Fail, messageArgs...)
			return err
		}

		if string(responseBuffer) == string(clientMessage) {
			logger.Error("check-response", logger.Fail, messageArgs...)
			return err
		}

		message, err := parseMessage(responseBuffer)
		if err != nil {
			logger.Error("parse-response", logger.Fail, messageArgs...)
			return err
		}

		if _, err := writer.Write(message); err != nil {
		    logger.Error("write-response-to-file", logger.Fail, messageArgs...)
		    return err
		}
		writer.Flush()

		messageId += 1 
		time.Sleep(ECHO_CLIENT_MESSAGE_DELAY_MS * time.Millisecond)
	}
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}
