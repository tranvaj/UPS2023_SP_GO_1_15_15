import time
from tkinter import messagebox
from typing import List
from IDispatchReceiver import IDispatchReceiver
from client_status import ClientStatus
from connection import TCPClient
from const import ClientMsgErr, ClientMsgOk, ClientMsgRecoveryLogin, MsgErrOpcode, MsgLoginOpcode


class LoginProcessor(IDispatchReceiver):
    def __init__(self, client_status: ClientStatus, tcp_client: TCPClient):
        self.client_status = client_status
        self.tcp_client = tcp_client
        self.tcp_client.server_ip = client_status.server_ip
        self.tcp_client.server_port = client_status.server_port
        self.tcp_client.connect()
        self.tcp_client.send_data(client_status.name, MsgLoginOpcode)
        self.data_received = False

    def receive_from_dispatcher(self, opcode, reply_status, reply_msg: List[str]):
        if opcode == MsgLoginOpcode:
            if reply_status == ClientMsgOk:
                self.client_status.logged_in = True
                self.client_status.board_size = int(reply_msg[1])
            elif reply_status == ClientMsgErr and reply_msg[0] == ClientMsgRecoveryLogin:
                self.client_status.recovery = True
                self.client_status.logged_in = True
                self.client_status.board_size = int(reply_msg[1])
            else:
                messagebox.showerror("Error", reply_msg[0])
        elif opcode == MsgErrOpcode:
            messagebox.showerror("Error", reply_msg[0])
        self.data_received = True

    def wait_for_data(self):
        currentTime = time.time()
        while not self.data_received:
            time.sleep(1)
            if time.time() - currentTime > 10:
                messagebox.showerror("Error", "Server did not respond. Closing.")
                exit()
        return self.data_received