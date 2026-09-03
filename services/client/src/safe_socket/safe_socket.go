package safe_socket

import "io"

func SendAll(socket io.Writer, bytes []byte) error {
	bytes_written := 0
	for bytes_written < len(bytes) {
		bytes_written_on_this_attempt, err := socket.Write(bytes[bytes_written:])
		if err != nil {
			return err
		} 
		bytes_written += bytes_written_on_this_attempt
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	bytes_readen = 0 
	for bytes_readen < size {
		bytes_readen_tmp, err := socket.Read(buff[bytes_readen:])
		if err != nil {
			return nil, err
		} 
		bytes_readen += bytes_readen_tmp
	}

	return buff, nil
}
