package client

import (
	"os"
	"io"
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

func (client *Client) formatMessage(message string) ([]byte, error) {
	messageLen := len(message) 

	if messageLen > 255 {	
		err := errors.New("integer overflow")
		logger.Error("format-message", logger.Fail, "message too long", err)
		return nil, err
	} 

	return append([]byte{uint8(messageLen)}, []byte(message)...), nil
}

func (client *Client) parseMessage(bytes []byte) ([]byte, error) {
	if len(bytes) < 1 {
		return nil, errors.New("empty buffer")
	}
	messageLen := int(bytes[0])
	if len(bytes) < 1+messageLen {	
    	return nil, errors.New("response buffer is shorter than message len")
	}
	return append(bytes[1 : 1+messageLen],'\r','\n'), nil
}

func (client *Client) sendMessage(message string, messageArgs []any) (error) {
	clientMessage, err := client.formatMessage(message)
	if err != nil {
		logger.Error("format-message", logger.Fail, messageArgs...)
		return err
	}

	if err := safe_socket.SendAll(client.conn, clientMessage); err != nil {
		logger.Error("send-message", logger.Fail, messageArgs...)
		return err
	}

	return nil
}

func (client *Client) receiveMessage(messageArgs []any) ([]byte, error) {
	var responseBuffer []byte
	var err error

	frameLenBytes, err := safe_socket.RecvAll(client.conn, 1)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil	// no message (no error)
		} else {
			logger.Error("", logger.Fail, err)
			return nil, err
		}
	}

	frameLen := uint8(frameLenBytes[0])
	if frameLen > 0 {
		responseBuffer, err = safe_socket.RecvAll(client.conn, int(frameLen))
		if err != nil {
			logger.Error("recv-message", logger.Fail, err)
			return nil, err
		}
	}
	return responseBuffer, nil
}

func (client *Client) requestReply(message string, messageId int) ([]byte, error) {
	messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", messageId}
	logger.Info("server-request-reply", logger.InProgress, messageArgs...)
		
	if err := client.sendMessage(message, messageArgs); err != nil {
		logger.Error("send-message", logger.Fail, messageArgs...)
		return nil, err
	}

	reply, err := client.receiveMessage(messageArgs)
	if err != nil {
		logger.Error("recv-message", logger.Fail, messageArgs...)
		return nil, err
	}

	return reply, nil
}

func (client *Client) Run() error {
	const mainAction = "test-echo-server"
	defer client.conn.Close()
	defer client.inputFile.Close()
	defer client.outputFile.Close()

	scanner := bufio.NewScanner(client.inputFile)
	writer := bufio.NewWriter(client.outputFile)
	
	responseBuffer, err := client.requestReply(client.config.AgencyId, -1)
	if responseBuffer != nil {
		logger.Error("protocol-error", logger.Fail, "Server respondio incorrectamente al enviar la agencia")
		return errors.New("Server respondio incorrectamente al enviar agencia")
    }	

	var messageArgs []any
	messageId := 0
	for scanner.Scan() {
		messageArgs = []any{"agency-id", client.config.AgencyId, "message-id", messageId}
		logger.Info(mainAction, logger.InProgress, messageArgs...)
		
		if responseBuffer, err := client.requestReply(scanner.Text(), messageId); err != nil {
			logger.Error("request-reply", logger.Fail, messageArgs...)
			return err
		} else if responseBuffer != nil {
			logger.Error("protocol-error", logger.Fail, "Server respondio incorrectamente al enviar la apuesta")
			return errors.New("Server respondio incorrectamente al enviar apuesta")
		}
		messageId += 1 
	}

	if 	tcpConn, ok := client.conn.(*net.TCPConn); ok {
        if err := tcpConn.CloseWrite(); err != nil {
            logger.Error("close-write", logger.Fail, "err", err)
            return err
        }
    }	
	
	// Recibir los ganadores 
	message, err := client.receiveMessage(messageArgs)
	for err == nil && message != nil {
	    outputFileLine := append(message, '\r', '\n')
	    if _, err = writer.Write(outputFileLine); err != nil {
	        logger.Error("write-response-to-file", logger.Fail, messageArgs...)
	        return err
	    }

	    message, err = client.receiveMessage(messageArgs)
	}

	if err != nil   {
	    logger.Error("receive-message", logger.Fail, messageArgs...)
	    return err
	}
	writer.Flush()
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}
