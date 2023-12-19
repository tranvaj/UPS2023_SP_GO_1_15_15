import time
from typing import List
from connection import TCPClient
from threading import Thread, Lock
from IDispatchReceiver import IDispatchReceiver

from const import ClientMsgOk, MaxNoPingReceived, MsgOkOpcode, MsgPingOpcode, MsgRecoveryOpcode, PingTime, RecoveryMaxAttempts

class Pinger(IDispatchReceiver):
    def __init__(self, tcp_client: TCPClient, game_gui, ping_retry_amount: int = MaxNoPingReceived, time_between_pings: int = PingTime, recover_periodically: bool = False):
        self.tcp_client = tcp_client
        self.connection_online = True
        self.ping_retry_amount = ping_retry_amount
        self.time_between_pings = time_between_pings
        self.game_gui = game_gui
        self.last_ping = time.time()
        self.last_recovery = time.time()
        self.recover_periodically = recover_periodically

    #ping server
    def __send_ping(self) -> bool:
        self.tcp_client.send_data("", MsgPingOpcode, ping=True)

    def __send_ping_thread(self):
        #periodically send ping to server
        while self.is_connection_online():
            self.__send_ping()
            time.sleep(self.time_between_pings)
            if self.timeSinceLastPing() > self.ping_retry_amount * self.time_between_pings:
                self.set_connection_online(False)
                self.game_gui.send_recovery(PingTime, RecoveryMaxAttempts)
                return
            

    def timeSinceLastPing(self):
        return time.time() - self.last_ping
    
    def timeSinceLastRecovery(self):
        return time.time() - self.last_recovery

    def receive_from_dispatcher(self, opcode, reply_status, reply_msg: List[str]):
        if opcode == MsgPingOpcode:
            if reply_status == ClientMsgOk:
                if self.recover_periodically and self.timeSinceLastRecovery() > PingTime*3:
                    self.tcp_client.send_data("", MsgRecoveryOpcode)
                    self.last_recovery = time.time()
                self.last_ping = time.time()

    def is_connection_online(self) -> bool:
        return self.connection_online
    
    def set_connection_online(self, connection_online: bool):
        self.connection_online = connection_online

    def run_pinger(self):
        #create own thread that periodically sends ping to server
        ping_thread = Thread(target=self.__send_ping_thread)
        ping_thread.daemon = True
        ping_thread.start()
