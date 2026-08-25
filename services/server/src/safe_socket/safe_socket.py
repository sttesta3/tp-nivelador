import socket

# TODO: Complete with a short-read/short-write tolerant implementation


def recv_all(socket: socket.socket, size):
    return socket.recv(size)


def send_all(socket: socket.socket, bytes):
    bytes_written = 0
    while bytes_written < len(bytes):
        bytes_written_on_this_attempt = socket.send(bytes[bytes_written:])
        if bytes_written_on_this_attempt < 1:
            # -1 = error. 0 = conexion cerrada ? 
            raise Exception
        bytes_written += bytes_written_on_this_attempt
    return bytes_written
