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
const CONNECTION_ATTEMPS_DELAY_MS = 1024

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	conn       net.Conn
	config 	   ClientConfig
	inputFile  *os.File
	outputFile *os.File
	state	   bool	
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

	client := &Client{conn: conn, config: config, inputFile: inputFile, outputFile: outputFile, state: true}
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

func (client *Client) formatFrame(message string) ([]byte, error) {
	messageLen := len(message) 

	if messageLen > 65535 {	
		err := errors.New("integer overflow")
		logger.Error("format-message", logger.Fail, "message too long", err)
		return nil, err
	} 

	// Big-endian Manual 
	highByte := byte(messageLen >> 8)
    lowByte := byte(messageLen & 0xFF)
	
	return append([]byte{highByte, lowByte}, []byte(message)...), nil
}

func (client *Client) sendMessage(message string) (error) {
	clientMessage, err := client.formatFrame(message)
	if err != nil {
		logger.Error("format-message", logger.Fail, "err", err)
		return err
	}

	if err := safe_socket.SendAll(client.conn, clientMessage); err != nil {
		logger.Error("send-message", logger.Fail, "err", err)
		return err
	}

	return nil
}

func (client *Client) receiveMessage() ([]byte, error) {
	action := "receive-message"
	var responseBuffer []byte
	var err error

	frameLenBytes, err := safe_socket.RecvAll(client.conn, 2)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil	// no hay mensaje (no hay error)
		} else {
			logger.Error(action, logger.Fail, "err", err)
			return nil, err
		}
	}	

	frameLen := (int(frameLenBytes[0]) << 8 | int(frameLenBytes[1]))
	if frameLen > 0 {
		responseBuffer, err = safe_socket.RecvAll(client.conn, frameLen)
		if err != nil {
			logger.Error("recv-message", logger.Fail, "err", err)
			return nil, err
		}
	}
	return responseBuffer, nil
}

func (client *Client) requestReply(message string, messageId int) ([]byte, error) {
	action := "request-reply"
	messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", messageId}
	logger.Info(action, logger.InProgress, messageArgs...)
		
	if err := client.sendMessage(message); err != nil {
		logger.Error(action, logger.Fail, messageArgs...)
		return nil, err
	}

	reply, err := client.receiveMessage()
	if err != nil {
		logger.Error(action, logger.Fail, messageArgs...)
		return nil, err
	}

	return reply, nil
}

func (client *Client) IsStop() bool {
	return ! client.state
} 


func (client *Client) Stop() {
	client.state = false
	if client.conn != nil {
		// Debemos cerrar el socket para desbloquear Recv / Send
		client.conn.Close()
	}
}

func (client *Client) Run() error {
	const mainAction = "test-echo-server"
	defer client.conn.Close()
	defer client.inputFile.Close()
	defer client.outputFile.Close()

	scanner := bufio.NewScanner(client.inputFile)
	writer := bufio.NewWriter(client.outputFile)
	
	if responseBuffer, err := client.requestReply(client.config.AgencyId, -1); err != nil {
		logger.Error(mainAction, logger.Fail, "err", err)
		return err
	} else if responseBuffer != nil {
		logger.Error("protocol-error", logger.Fail, "err", "Server respondio incorrectamente al enviar la agencia")
		return errors.New("Server respondio incorrectamente al enviar agencia")
    }	

	var messageArgs []any
	messageId := 0
	
	moreLines := scanner.Scan()
	for moreLines && client.state {
		messageArgs = []any{"agency-id", client.config.AgencyId, "message-id", messageId}
		logger.Info(mainAction, logger.InProgress, messageArgs...)
		
		betsInMessage := 0 
		chunkMessage := ""
		fullMessage := false
		for ! fullMessage && client.state {
			line := scanner.Text()
			if moreLines && len(line)+len(chunkMessage) <= 65534 && betsInMessage < client.config.BatchSize {
				chunkMessage += line + "\n"
				betsInMessage += 1 
				moreLines = scanner.Scan()
			} else {
				fullMessage = true
			}
		}

		if client.state {
			if responseBuffer, err := client.requestReply(chunkMessage, messageId); err != nil {
				logger.Error("request-reply", logger.Fail, messageArgs...)
				return err
			} else if responseBuffer != nil {
				logger.Error("protocol-error", logger.Fail, "err", "Server respondio incorrectamente al enviar la apuesta")
				return errors.New("Server respondio incorrectamente al enviar apuesta")
			}
		}

		messageId += 1 
	}

	if client.state {
		if 	tcpConn, ok := client.conn.(*net.TCPConn); ok {
        	if err := tcpConn.CloseWrite(); err != nil {
    	        logger.Error("close-write", logger.Fail, "err", err)
	            return err
    	    }
	    }	
	}
	
	// Recibir los ganadores (si se llega a este punto, se deja terminar) 
	if client.state {
		message, err := client.receiveMessage()
		for err == nil && message != nil {
	    	outputFileLine := append(message, '\r', '\n')
		    if _, err = writer.Write(outputFileLine); err != nil {
		        logger.Error("write-response-to-file", logger.Fail, "err", err)
	        	return err
	    	}

	    	message, err = client.receiveMessage()
		}

		if err != nil   {
	    	logger.Error("receive-message", logger.Fail, messageArgs...)
		    return err
		}
	}
	writer.Flush()
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)

	return nil
}
