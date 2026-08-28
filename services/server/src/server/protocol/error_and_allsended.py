from .messages import Message, ALL_SENDED, ERROR
from __future__ import annotations

class AllSendedMessage(Message):
    def __init__(self, agency_id: int):
        self.type = ALL_SENDED
        super().__init__(agency_id)

    def _marshall_payload(self) -> bytes:
        return b""

    @classmethod
    def _unmarshall_payload(cls, agency_id: int, payload: bytes) -> AllSendedMessage:
        return cls(agency_id)



class ErrorMessage(Message):
    def __init__(self, agency_id: int, message: str = ""):
        self.type = ERROR
        super().__init__(agency_id)
        self.message = message
    def __init__(self, agency_id: int, message: str = ""):
        super().__init__(agency_id)
        self.message = message

    def _marshall_payload(self) -> bytes:
        return self.message.encode('utf-8')

    @classmethod
    def _unmarshall_payload(cls, agency_id: int, payload: bytes) -> ErrorMessage:
        return cls(agency_id, payload.decode('utf-8'))