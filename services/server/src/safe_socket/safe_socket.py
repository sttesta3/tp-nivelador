import socket

# TODO: Complete with a short-read/short-write tolerant implementation


def recv_all(socket: socket.socket, size):
    # Misma logica que codigo en cliente, para mas informacion leer comentario en safe_socket cliente
    bytes_readen = socket.recv(size)
    if bytes_readen:
        while len(bytes_readen) < bytes_readen[0] + 1:
            bytes_readen += socket.recv(size)

    return bytes_readen

def send_all(socket: socket.socket, bytes):
    bytes_written = 0
    while bytes_written < len(bytes):
        bytes_written_on_this_attempt = socket.send(bytes[bytes_written:])
        if bytes_written_on_this_attempt < 1:
            raise Exception("Eh boca boca")
        bytes_written += bytes_written_on_this_attempt
    return bytes_written
