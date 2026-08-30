import socket

def recv_all(socket, size):
    bytes_received = b''
    while len(bytes_received) < size:
        chunk = socket.recv(size - len(bytes_received))
        if not chunk:
            raise ConnectionError("closed before receiving all data")
        bytes_received += chunk
    return bytes_received

def send_all(socket: socket.socket, bytes):
    #  this method continues to send data from bytes until either all data has been sent or an error occurs
    return socket.sendall(bytes)
