import socket
import logger

def recv_all(socket: socket.socket, size):
    bytes_readen = bytearray()
    while len(bytes_readen) < size:  
        bytes_readen_tmp = socket.recv(size - len(bytes_readen))
        if not bytes_readen_tmp:
            return
        bytes_readen += bytes_readen_tmp 
    return bytes_readen

def send_all(socket: socket.socket, bytes):
    bytes_written = 0
    while bytes_written < len(bytes):
        bytes_written_on_this_attempt = socket.send(bytes[bytes_written:])
        if bytes_written_on_this_attempt < 1:
            raise Exception("[safe_socket][send_all] bytes_written_on_this_attempt es menor a 1")
        bytes_written += bytes_written_on_this_attempt
    return bytes_written
