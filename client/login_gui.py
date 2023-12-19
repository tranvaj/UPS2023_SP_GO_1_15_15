import tkinter as tk
from tkinter import messagebox
from typing import List
from connection import TCPClient
from const import ArgSep, ClientMsgErr, ClientMsgOk, ClientMsgRecoveryLogin, ConnHost, ConnPort, MsgErrOpcode, MsgMagic, MsgLoginOpcode, MsgOkOpcode
from game_gui import TicTacToeGUI
from client_status import ClientStatus
import threading

class LoginScreen(tk.Tk):
    def __init__(self, client_status: ClientStatus):
        try:
            super().__init__()
            self.title("Login")
            self.geometry("300x300")
            self.resizable(False, False)
            self.client_status = client_status

            self.username_frame = tk.Frame(self)
            self.username_frame.pack(fill=tk.BOTH, expand=True)

            self.username_label = tk.Label(self.username_frame, text="Username:", padx=10, pady=5)
            self.username_label.pack()

            self.username_entry = tk.Entry(self.username_frame)
            self.username_entry.pack(pady=(0,10))  # Add padding to the entry widget

            self.ip_frame = tk.Frame(self)
            self.ip_frame.pack(fill=tk.BOTH, expand=True)

            self.ip_label = tk.Label(self.ip_frame, text="IP:", padx=10, pady=5)
            self.ip_label.pack()

            self.ip_entry = tk.Entry(self.ip_frame)
            self.ip_entry.insert(tk.END, ConnHost)  # Set default IP address
            self.ip_entry.pack(pady=(0,10))  

            self.port_frame = tk.Frame(self)
            self.port_frame.pack(fill=tk.BOTH, expand=True)

            self.port_label = tk.Label(self.port_frame, text="Port:", padx=10, pady=5)
            self.port_label.pack()

            self.port_entry = tk.Entry(self.port_frame)
            self.port_entry.insert(tk.END, ConnPort)  # Set default port number
            self.port_entry.pack(pady=(0,10)) 

            self.login_button = tk.Button(self, text="Connect", command=self.login, padx=10, pady=5)
            self.login_button.pack(fill=tk.X)

            self.protocol("WM_DELETE_WINDOW", self.on_closing)
        except Exception as e:
            messagebox.showerror("Error", str(e))
            self.destroy()
            return

    def on_closing(self):
        if messagebox.askokcancel("Quit", "Do you want to quit?"):
            self.destroy()

    # def receive_from_dispatcher(self, opcode, reply_status, reply_msg: List[str]):
    #     if opcode == MsgLoginOpcode:
    #         if reply_status == ClientMsgOk:
    #             self.client_status.logged_in = True
    #             self.client_status.name = self.username_entry.get()
    #             self.client_status.board_size = int(reply_msg[1])
    #         elif reply_status == ClientMsgErr and reply_msg[0] == ClientMsgRecoveryLogin:
    #             self.client_status.recovery = True
    #             self.client_status.logged_in = True
    #             self.client_status.name = self.username_entry.get()
    #             self.client_status.board_size = int(reply_msg[1])
    #         else:
    #             messagebox.showerror("Error", reply_msg[0])
    #         self.destroy()
    #     elif opcode == MsgErrOpcode:
    #         messagebox.showerror("Error", reply_msg[0])
    #         self.destroy()

    def login(self):
        #username = self.username_entry.get()
        #self.tcp_client.send_data(username, MsgLoginOpcode)
        self.client_status.name = self.username_entry.get()
        self.client_status.server_ip = self.ip_entry.get()
        self.client_status.server_port = int(self.port_entry.get())
        self.destroy()





