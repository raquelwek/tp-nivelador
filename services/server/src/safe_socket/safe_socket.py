import socket
import logger
def recv_all(socket, size):
    logger.info("recv_all", logger.LogResult.in_progress, "waiting", "bytes", size)
    bytes_received = b''
    while len(bytes_received) < size:
        chunk = socket.recv(size - len(bytes_received))
        logger.info("recv_all", logger.LogResult.in_progress, "received", "bytes", len(chunk))
        if not chunk:
            raise ConnectionError("closed before receiving all data")
        bytes_received += chunk
    logger.info("recv_all", logger.LogResult.success, "received", "total", len(bytes_received))
    return bytes_received

def send_all(socket: socket.socket, bytes):
    #  this method continues to send data from bytes until either all data has been sent or an error occurs
    return socket.sendall(bytes)
