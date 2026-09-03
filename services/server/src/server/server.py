import socket
import logger
import threading
import os
import signal

from server.protocol.messages import HEADER_LENGTH, unmarshall_message, Message, ALL_SENDED, BETS
from server.protocol.bets_records import WinnersMessage
from server.protocol.simple_messages import ErrorMessage, AckMessage
from lottery.lottery import Lottery
from safe_socket.safe_socket import recv_all, send_all
from server.rw_lock import RWLock

BETS_RECEIVED_NAME_FILE = "bets_received.csv"
SOCKET_TIMEOUT_ACCEPT = 1.0
SHUTDOWN_GRACE_PERIOD = 5.0

class Server:
    def __init__(self, server_host: str, server_port: int, agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port

        self._file_lock = RWLock()
        self.agency_quorum_min = agency_quorum_min
        self._quorum_barrier = threading.Barrier(agency_quorum_min)  # for managing the quorum

        # for graceful shutdown
        self._running = True
        signal.signal(signal.SIGTERM, self._handle_shutdown)

        # Crear el archivo si no existe
        if not os.path.exists(BETS_RECEIVED_NAME_FILE):
            with open(BETS_RECEIVED_NAME_FILE, "w") as _:
                pass 

        self.lottery = Lottery(BETS_RECEIVED_NAME_FILE)

    '''
    Contains the logic responses and message meaning definited in the protocol
    '''
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
                    logger.info(action, logger.LogResult.in_progress, "waiting-for-quorum", "agency-id", agency_id)
                    try: 
                        self._quorum_barrier.wait()
                    except threading.BrokenBarrierError:
                        logger.error(action, logger.LogResult.fail, "quorum-aborted", "agency-id", agency_id)
                        return
                    
                    logger.info(action, logger.LogResult.success, "quorum-reached", "agency-id", agency_id)
                    winners_message = self.getWinnersMessage(agency_id)
                    self.send_message(client_socket, winners_message)
                    logger.info(action, logger.LogResult.success, "sent", "winners", "message", str(winners_message))
                    logger.info(action, logger.LogResult.success, "messages-amount", message_amount, "agency-id", agency_id)
                    return 
                
                elif client_message.type == BETS:
                    self.handle_bets_message(client_message)
                    self.send_ack(client_socket, agency_id)

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
                    logger.error(action, logger.LogResult.fail, "error", str(e))
                    break
                
                logger.info(action, logger.LogResult.success)
                thread = threading.Thread(target=self._handle_client, args=(client_socket,))
                thread.start()
                threads.append((client_socket,thread))


        self._graceful_shutdown(threads)
        logger.info("shutdown", logger.LogResult.success)

    '''
    Closes all client sockets and waits for all threads to finish. It also aborts the quorum barrier to unblock any waiting threads.
    If the client has finished and closed the conection, it continues with the ones that havent
    '''
    def _graceful_shutdown(self, threads):
        action = "shutdown"
        logger.info(action, logger.LogResult.in_progress, "initiating-graceful-shutdown")
        self._quorum_barrier.abort()

        for client_socket, _ in threads: 
            try: 
                client_socket.shutdown(socket.SHUT_RDWR)
            except OSError: pass
            try:
                client_socket.close()
            except OSError: pass

        for _, thread in threads:
            thread.join()

        
        
    
    '''
    Receives the bytes of a message following the protocol and unmarshalls it into a Message object.
    It first receives the header, then the payload, and finally combines them to create the message.
    '''
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

    '''
    Sends a message to the client socket
    '''
    def send_message(self, client_socket: socket.socket, message: Message):
        action = "send-message"
        message_bytes = message.marshall()
        send_all(client_socket, message_bytes)
        logger.info(action, logger.LogResult.success, "message", str(message))
    '''
    Stores the bests received in the message
    '''
    def handle_bets_message(self, message: Message):
        action = "handle-bets-message"
        logger.info(action, logger.LogResult.in_progress, "bets-count", len(message.bets))
        with self._file_lock.write_lock():
            self.lottery.store_bets(message.bets)
        logger.info(action, logger.LogResult.success, "bets-count", len(message.bets))

    '''
    Returns a Winners message with all the winners of the agency_id passed as parameter. If there are no winners, it returns an empty Winners message.
    '''
    def getWinnersMessage(self, agency_id: int) -> Message:
        action = "get-winners"
        winners_message = WinnersMessage(agency_id)
        logger.info(action, logger.LogResult.in_progress, "agency-id", agency_id)
        with self._file_lock.read_lock():
            for bet in self.lottery.load_bets():
                if not (self.lottery.has_won(bet) and bet.agency_id == agency_id):continue
                winners_message.add_bet(bet)
        
        logger.info(action, logger.LogResult.success, "winners-count", len(winners_message.bets), "agency-id", agency_id)
        return winners_message
    '''
    Handles the shutdown signal (SIGTERM) and sets the running flag to False
    in order to stop the server gracefully. It also logs the shutdown event.
    '''
    def _handle_shutdown(self, sig, frame):
        logger.info("shutdown", logger.LogResult.in_progress)
        self._running = False

    def send_ack(self, client_socket: socket.socket, agency_id: int):
        ack_message = AckMessage(agency_id)
        self.send_message(client_socket, ack_message)
       