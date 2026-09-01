from __future__ import annotations
from abc import ABC, abstractmethod
BETS = 0x01
ALL_SENDED = 0x02
WINNERS = 0x03
ERROR = 0x04
ACK = 0x05

HEADER_LENGTH = 6  # bytes


class Message(ABC):
    type: int  # lo fija el decorador @register_message en cada subclase

    def __init__(self, agency_id: int, batch_size: int = 1024):
        self.agency_id = agency_id
        self.batch_size = batch_size

    def marshall(self) -> bytes:
        """Template method: arma el header y delega el payload a la subclase."""
        payload = self._marshall_payload()
        header = self.type.to_bytes(1, byteorder='big')
        header += self.agency_id.to_bytes(1, byteorder='big')
        header += len(payload).to_bytes(4, byteorder='big')  # <- siempre derivado, nunca acumulado a mano
        return header + payload

    @abstractmethod
    def _marshall_payload(self) -> bytes:
        """Cada subclase serializa SU payload."""
        raise NotImplementedError

    @classmethod
    @abstractmethod
    def _unmarshall_payload(cls, agency_id: int, payload: bytes) -> Message:
        """Cada subclase reconstruye una instancia propia a partir del payload crudo."""
        raise NotImplementedError


def unmarshall_message(data: bytes) -> Message:
    # Import here to avoid circular imports at module load time
    from .bets_records import BetsMessage, WinnersMessage
    from .simple_messages import AllSendedMessage, ErrorMessage, AckMessage
    
    if len(data) < HEADER_LENGTH:
        raise ValueError("Data too short to be a valid Message")

    type_id = data[0]
    agency_id = data[1]
    payload_length = int.from_bytes(data[2:6], byteorder='big')

    if len(data) < HEADER_LENGTH + payload_length:
        raise ValueError("Data too short for the specified payload length")

    payload = data[HEADER_LENGTH:HEADER_LENGTH + payload_length]

    if type_id == BETS:
        return BetsMessage._unmarshall_payload(agency_id, payload)
    elif type_id == ALL_SENDED:
        return AllSendedMessage._unmarshall_payload(agency_id, payload)
    elif type_id == WINNERS:
        return WinnersMessage._unmarshall_payload(agency_id, payload)
    elif type_id == ERROR:
        return ErrorMessage._unmarshall_payload(agency_id, payload)
    elif type_id == ACK:
        return AckMessage._unmarshall_payload(agency_id, payload)
    else:
        raise ValueError(f"Unknown message type: {type_id}")