import os
import sys
import threading
from IDispatchReceiver import IDispatchReceiver
from connection import TCPClient
from const import ArgSep, MsgContinueOpcode, MsgGameOverOpcode, MsgGameStartedOpcode, MsgJoinOpcode, MsgLoginOpcode, MsgMoveOpcode, MsgPauseOpcode, MsgPingOpcode, MsgPlayAgainOpcode, MsgRecoveryOpcode, MsgReturnToStartOpcode, MsgYourTurnOpcode
from game_gui import TicTacToeGUI
import select

class Dispatcher:
    def __init__(self, tcp_client: TCPClient, game_gui: IDispatchReceiver, login: IDispatchReceiver, pinger: IDispatchReceiver):
        self.tcp_client = tcp_client
        self.game_gui = game_gui
        self.login = login
        self.pinger = pinger

    def send_data(self, data: str, opcode: str):
        self.tcp_client.send_data(data, opcode)

    def __receive_data(self) -> str:
        data = self.tcp_client.receive_data()
        opcode = data[1]
        reply = data[0].split(ArgSep)
        reply_status = reply[0]
        reply_msg = reply[1:]
        
        if self.login == None:
            raise Exception("login gui doesn't exist")      
        elif opcode == MsgLoginOpcode:
            self.login.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif self.game_gui == None:
            raise Exception("game gui doesn't exist")
        elif opcode == MsgJoinOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif opcode == MsgMoveOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif opcode == MsgReturnToStartOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif opcode == MsgGameOverOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif opcode == MsgYourTurnOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif opcode == MsgPlayAgainOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg) 
        elif opcode == MsgGameStartedOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif opcode == MsgRecoveryOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif opcode == MsgPingOpcode:
            self.pinger.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif opcode == MsgPauseOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg)
        elif opcode == MsgContinueOpcode:
            self.game_gui.receive_from_dispatcher(opcode, reply_status, reply_msg)
            
            
        return data
    
    def set_game_gui(self, game_gui: IDispatchReceiver):
        self.game_gui = game_gui

    def set_login_gui(self, login_gui: IDispatchReceiver):
        self.login = login_gui

    def set_pinger(self, pinger: IDispatchReceiver):
        self.pinger = pinger
    
    def run_dispatcher(self):
        dispatcher_thread = threading.Thread(target=self.listen)
        dispatcher_thread.daemon = True
        dispatcher_thread.start()

    def listen(self):
        #use select to listen for data from server
        while True:
            try:
                ready = select.select([self.tcp_client.client_socket], [], [], None)
                if ready[0]:
                    self.__receive_data()
            except Exception as e:
                print("Error: " + str(e))
                print("Exiting due to invalid receive...")
                os._exit(1)