package client

import (
	"os"
	"net"
	"time"
	"bufio"

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
	conn   net.Conn
	config ClientConfig
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
		
		clientMessage := scanner.Text()

		if err := safe_socket.SendAll(client.conn, []byte(clientMessage)); err != nil {
			logger.Error("send-message", logger.Fail, messageArgs...)
			return err
		}

		responseBuffer, err := safe_socket.RecvAll(client.conn, ECHO_CLIENT_BUFFER_SIZE)
		if err != nil {
			logger.Error("recv-response", logger.Fail, messageArgs...)
			return err
		}

		if string(responseBuffer) == clientMessage {
			logger.Error("check-response", logger.Fail, messageArgs...)
			return err
		}

		// make crea un buffer inicializado en cero. 
		// se debe contar la cantidad de bytes distintos de null para no escribir basura
		valid_byte_count := 0
		for responseBuffer[valid_byte_count] != '\x00' {
			valid_byte_count += 1
		}
		responseBuffer[valid_byte_count] = '\r'
		responseBuffer[valid_byte_count+1] = '\n'
		valid_byte_count += 2 

		writen_bytes, err := writer.Write(responseBuffer[0:valid_byte_count])
		if writen_bytes != valid_byte_count || err != nil {
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
