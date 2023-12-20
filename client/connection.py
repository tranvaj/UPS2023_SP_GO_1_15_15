import socket
import time

from const import MsgHeaderLen, MsgMagic, MsgLoginOpcode
from message_formatter import log_message

class TCPClient:
    def __init__(self, server_ip, server_port):
        self.server_ip = server_ip
        self.server_port = server_port
        self.client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.block_send = False

    def connect(self):
        self.client_socket.connect((self.server_ip, self.server_port))

    
    def send_data(self, data, opcode, ping: bool = False):
        """
        Send data to server
        Throws exception if not sent successfully
        """

        if self.block_send:
            return
        padded_length = str(len(data)).zfill(4)
        msg = f"{MsgMagic}{opcode}{padded_length}{data}"
        if ping:
            try:
                self.client_socket.sendall(bytes(msg, "utf-8"))
                print(log_message("Sent ping"))
            except Exception as e:
                print("Could not send ping: " + str(e))
        else:
            try:
                self.client_socket.sendall(bytes(msg, "utf-8"))
            except Exception as e:
                print(log_message("Could not send data: " + str(e)))
    def set_block_send(self, block_send):
        self.block_send = block_send
        
    def receive_data(self):
        # Receive header first
        msg = str(self.client_socket.recv(MsgHeaderLen), "utf-8")
        print(log_message(f"Received header: {msg}"))

        # Get magic
        magic = msg[:len(MsgMagic)]
        if magic != MsgMagic:
            raise ValueError("Magic does not match!")

        # Get opcode
        opcode = msg[len(MsgMagic):len(MsgMagic) + len(MsgLoginOpcode)]
        if len(opcode) != 3:
            raise ValueError("Invalid opcode format!")

        # Get data length
        data_len = msg[len(MsgMagic) + len(MsgLoginOpcode):]
        if len(data_len) != 4:
            raise ValueError("Invalid data length format!")

        data_len = int(data_len)

        # Read data
        data = self.client_socket.recv(data_len)
        data = data.decode("utf-8")

        print(log_message("Received data:" + data))
        return (data, opcode)
    
    def set_timeout(self, timeout):
        self.client_socket.settimeout(timeout)

    def reconnect(self):
        self.client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try: 
            self.client_socket.connect((self.server_ip, self.server_port))
        except Exception as e:
            print("SOMETHING WRONG: " + str(e))
            raise e

    def close(self):
        self.client_socket.close()
        print("Connection closed.")


