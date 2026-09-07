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
	bytes_readen, err := socket.Read(buff)
	if err != nil {
		return nil, err
	} else if bytes_readen > 0 {	// Si no recibo bytes en primer lectura, me cerraron el socket devuelvo cero
		for bytes_readen < size {
			bytes_readen_tmp, errIn := socket.Read(buff[bytes_readen:])
			if errIn != nil {
				return nil, errIn
			} else if bytes_readen_tmp == 0 {	// No recibi bytes de un socket bloqueante. Es EOF prematuro
				return buff[:bytes_readen], io.EOF
			} 
			bytes_readen += bytes_readen_tmp
		}
	} 

	return buff, nil
}
