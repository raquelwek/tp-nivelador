from __future__ import annotations
from .messages import Message, BETS, WINNERS
from lottery.bet import Bet

LEN_BET_MIN = 16


class BetsMessage(Message):
    def __init__(self, agency_id: int, batch_size: int = 1024):
        self.type = BETS
        super().__init__(agency_id, batch_size)
        self.bets = []

    def add_bet(self, bet):
        if len(self.bets) >= self.batch_size:
            raise ValueError("cannot add bet: batch is full")
        self.bets.append(bet)

    def _marshall_payload(self) -> bytes:
        payload = len(self.bets).to_bytes(2, byteorder='big')
        for bet in self.bets:
            payload += bet.document.to_bytes(4, byteorder='big')
            payload += bet.number.to_bytes(2, byteorder='big')
            name_bytes = bet.first_name.encode('utf-8')
            last_name_bytes = bet.last_name.encode('utf-8')
            payload += len(name_bytes).to_bytes(1, byteorder='big')
            payload += len(last_name_bytes).to_bytes(1, byteorder='big')
            payload += bet.birthdate.encode('utf-8').ljust(8, b'\x00')
            payload += name_bytes
            payload += last_name_bytes
        return payload

    @classmethod
    def _unmarshall_payload(cls, agency_id: int, payload: bytes) -> BetsMessage:
        msg = cls(agency_id)
        bets_count = int.from_bytes(payload[0:2], byteorder='big')
        offset = 2
        for _ in range(bets_count):
            document = int.from_bytes(payload[offset:offset + 4], byteorder='big'); offset += 4
            number = int.from_bytes(payload[offset:offset + 2], byteorder='big'); offset += 2
            name_length = payload[offset]; offset += 1
            last_name_length = payload[offset]; offset += 1
            birthdate = payload[offset:offset + 8].decode('utf-8').rstrip('\x00'); offset += 8
            name = payload[offset:offset + name_length].decode('utf-8'); offset += name_length
            last_name = payload[offset:offset + last_name_length].decode('utf-8'); offset += last_name_length

            bet = Bet(agency_id, name, last_name, document, birthdate, number)
            msg.bets.append(bet)  # ojo: append directo, NO add_bet (ver nota abajo)
        return msg

class WinnersMessage(BetsMessage):
    type = WINNERS