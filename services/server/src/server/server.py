import socket
import logger
import safe_socket
import threading
import time
import os
import signal

from server.protocol.messages import HEADER_LENGTH, unmarshall_message, Message, ALL_SENDED, BETS
from server.protocol.bets_records import WinnersMessage
from server.protocol.error_and_allsended import ErrorMessage
from lottery.lottery import Lottery
from safe_socket.safe_socket import recv_all, send_all

BETS_RECEIVED_NAME_FILE = "bets_received.csv"
SOCKET_TIMEOUT_ACCEPT = 1.0

class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        # for graceful shutdown
        self._running = True
        signal.signal(signal.SIGTERM, self._handle_shutdown)

        # Crear el archivo si no existe
        if not os.path.exists(BETS_RECEIVED_NAME_FILE):
            with open(BETS_RECEIVED_NAME_FILE, "w") as _:
                pass  # Crear archivo vacío

        self.lottery = Lottery(BETS_RECEIVED_NAME_FILE)

    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0

        try:
            logger.info(action, logger.LogResult.in_progress)
            while True:         
                client_message = self.recv_message(client_socket)

                agency_id = client_message.agency_id

                logger.info(action, logger.LogResult.in_progress, "received", "message", str(client_message), "messages-amount",
                                        message_amount, "agency-id", agency_id)
                message_amount += 1

                if client_message.type == ALL_SENDED:  
                    # @TO DO: Manage concurrency to wait until AGENCY_QUORUM_MIN is reached
                    winners_message = self.getWinnersMessage(agency_id)
                    self.send_message(client_socket, winners_message)
                    logger.info(action, logger.LogResult.success, "sent", "winners", "message", str(winners_message))
                    logger.info(action, logger.LogResult.success, "messages-amount", message_amount, "agency-id", agency_id)
                    return 
                
                elif client_message.type == BETS:
                    self.handle_bets_message(client_message)

        except Exception as e:
            logger.error(action, logger.LogResult.fail, "error", str(e))
        finally:
            client_socket.close()

    def run(self):
        action = "run-server"
        logger.info(action, logger.LogResult.in_progress, "host", self.server_host, "port", self.server_port)
        action = "accept-connection"
        threads = []
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            server_socket.settimeout(SOCKET_TIMEOUT_ACCEPT)

            while self._running:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except socket.timeout:
                    continue

                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    break
                
                logger.info(action, logger.LogResult.success)

                thread = threading.Thread(target=self._handle_client, args=(client_socket,))
                thread.start()
                threads.append(thread)


        for thread in threads:
            thread.join()
        logger.info("shutdown", logger.LogResult.success)

    def recv_message(self, client_socket:socket.socket) -> Message:
        action = "recv-message"
        logger.info(action, logger.LogResult.in_progress, "waiting", "header")
        header_bytes = recv_all(client_socket, HEADER_LENGTH)
        logger.info(action, logger.LogResult.in_progress, "received", "header", "bytes", len(header_bytes))
        payload_length = int.from_bytes(header_bytes[2:6], byteorder='big')
        logger.info(action, logger.LogResult.in_progress, "waiting", "payload", "length", payload_length)
        payload_bytes = recv_all(client_socket, payload_length)
        logger.info(action, logger.LogResult.in_progress, "received", "payload", "bytes", len(payload_bytes))
        message_bytes = header_bytes + payload_bytes
        message = unmarshall_message(message_bytes)
        logger.info(action, logger.LogResult.success, "message", str(message))
        return message

    def send_message(self, client_socket: socket.socket, message: Message):
        action = "send-message"
        message_bytes = message.marshall()
        send_all(client_socket, message_bytes)
        logger.info(action, logger.LogResult.success, "message", str(message))

    def handle_bets_message(self, message: Message):
        action = "handle-bets-message"
        self.lottery.store_bets(message.bets)
        logger.info(action, logger.LogResult.success, "bets-count", len(message.bets))

    '''
    Returns a Winners message with all the winners of the agency_id passed as parameter. If there are no winners, it returns an empty Winners message.
    '''
    def getWinnersMessage(self, agency_id: int) -> Message:
        action = "get-winners"
        winners_message = WinnersMessage(agency_id)
        winners = 0
        for bet in self.lottery.load_bets():
            if not (self.lottery.has_won(bet) and bet.agency_id == agency_id):
                continue
            winners_message.add_bet(bet)
            winners += 1

        logger.info(action, logger.LogResult.success, "winners-count", winners)
        return winners_message

    def _handle_shutdown(self, signum, frame):
        logger.info("shutdown", logger.LogResult.in_progress)
        self._running = False