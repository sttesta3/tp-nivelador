package safe_socket

import "io"

//TODO: Complete with a short-read/short-write tolerant implementation

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
	_, err := socket.Read(buff)
	if err != nil {
		return nil, err
	}
	return buff, nil
}
