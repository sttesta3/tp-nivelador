package safe_socket

import (
	"io"
	"errors"
)

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
	// Logica propuesta: 
	// Leer tamaño del buffer. El primer byte es el tamaño del mensaje. Los siguientes el contenido
	// Leer mientras la cantidad de bytes leidos sea menor a dos bytes. Luego leer hasta cumplir con el tamaño anterior
	// Nota: Se utiliza este metodo asumiendo que es un protocolo request-reply. No se enviaran mensajes seguidos 

	buff := make([]byte, size)
	bytes_readen, err := socket.Read(buff)
	if err != nil {
		return nil, err
	} else if bytes_readen == 0 {
		return nil, errors.New("connection closed")
	}

	for bytes_readen < int(buff[0]) + 1 {
		bytes_readen_tmp, err := socket.Read(buff[bytes_readen:])
		if err != nil {
			return nil, err
		} else if bytes_readen == 0 {
			return nil, errors.New("connection closed")
		}
		bytes_readen += bytes_readen_tmp
	}

	return buff, nil
}
