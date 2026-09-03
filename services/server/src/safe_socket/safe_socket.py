import socket

def recv_all(socket, size):
    bytes_received = b''
    while len(bytes_received) < size:
        chunk = socket.recv(size - len(bytes_received))
        if not chunk: 
            # chunk == b'' indicates that the connection has been closed
            raise ConnectionError("closed before receiving all data")
        bytes_received += chunk
    return bytes_received

def send_all(socket: socket.socket, bytes):
    bytes_sent = 0
    while bytes_sent < len(bytes):
        sent = socket.send(bytes[bytes_sent:])
        bytes_sent += sent
        
    return bytes_sent
