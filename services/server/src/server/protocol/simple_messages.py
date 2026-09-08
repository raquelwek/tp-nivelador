from __future__ import annotations
from .messages import Message, ALL_SENDED, ACK

class AllSendedMessage(Message):
    def __init__(self, agency_id: int):
        self.type = ALL_SENDED
        super().__init__(agency_id)

    def _marshall_payload(self) -> bytes:
        return b""

    @classmethod
    def _unmarshall_payload(cls, agency_id: int, payload: bytes) -> AllSendedMessage:
        return cls(agency_id)


class AckMessage(Message):
    def __init__(self, agency_id: int):
        self.type = ACK
        super().__init__(agency_id)

    def _marshall_payload(self) -> bytes:
        return b""

    @classmethod
    def _unmarshall_payload(cls, agency_id: int, payload: bytes) -> AckMessage:
        return cls(agency_id)