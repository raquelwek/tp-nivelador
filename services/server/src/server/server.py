import socket
import logger
import safe_socket

from server.protocol.messages import HEADER_LENGTH, unmarshall_message, Message, ALL_SENDED,WinnersMessage, ErrorMessage, BETS
from server.src_frozen.lottery import Lottery

BETS_RECEIVED_NAME_FILE = "bets_received.csv"

class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery(BETS_RECEIVED_NAME_FILE)

    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0
        try:
            logger.info(action, logger.LogResult.in_progress)
            while True:
                client_message = self.recv_message(client_socket)
                
                if not client_message:
                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        message_amount,
                    )
                    return
                message_amount += 1

                if client_message.type == ALL_SENDED:  
                    winners_message = WinnersMessage(client_message.agency_id)
                    self.send_message(client_socket, winners_message)
                elif client_message.type == BETS:
                    self.handle_bets_message(client_message)

        except Exception as e:
            logger.error(action, logger.LogResult.fail)
            self.send_message(client_socket, ErrorMessage(client_message.agency_id, str(e)))
                    

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)

    def recv_message(self, client_socket: socket.socket) -> Message:
        action = "recv-message"
        header_bytes = client_socket.recv_all(HEADER_LENGTH)
        payload_length = int.from_bytes(header_bytes[2:6], byteorder='big')
        payload_bytes = client_socket.recv_all(payload_length)
        message_bytes = header_bytes + payload_bytes
        message = unmarshall_message(message_bytes)
        logger.info(action, logger.LogResult.success, "message", str(message))
        return message

    def send_message(self, client_socket: socket.socket, message: Message):
        action = "send-message"
        message_bytes = message.marshall()
        client_socket.send_all(message_bytes)
        logger.info(action, logger.LogResult.success, "message", str(message))

    def 
