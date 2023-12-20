import datetime
import threading
import time
import tkinter as tk
from message_formatter import log_message
from typing import List
from IDispatchReceiver import IDispatchReceiver
from client_status import ClientStatus

from connection import TCPClient
from const import ClientMsgErr, ClientMsgGameGone, ClientMsgOk, ClientMsgRecovery_InGame_GameOver, ClientMsgRecovery_InGame_OtherPlayAgain, ClientMsgRecovery_InGame_OtherTurn, ClientMsgRecovery_InGame_YourTurn, ClientMsgRecovery_InLobby, ClientMsgRecovery_ReadyForGame, MsgContinueOpcode, MsgLoginOpcode, MsgPauseOpcode, MsgRecoveryOpcode, MsgStatusOpcode, PingTime, RecoveryMaxAttempts, rowSep, colSep, MsgGameOverOpcode, MsgGameStartedOpcode, MsgJoinOpcode, MsgMoveOpcode, MsgOkOpcode, MsgPlayAgainOpcode, MsgReturnToStartOpcode, MsgYourTurnOpcode
from pinger import Pinger

class TicTacToeGUI(IDispatchReceiver):
    def __init__(self, client_status: ClientStatus, tcp_client: TCPClient, root=None):
        self.client_status = client_status
        self.tcp_client = tcp_client
        self.pinger = Pinger(tcp_client=tcp_client, game_gui=self)
        self.recovery_finished = False
        self.recovery_mutex = threading.Lock()
        #threading.Thread(target=self.pinger.run_pinger).start()

        if root is None:
            self.root = tk.Tk()
        else:
            self.root = root
        self.root.title("Tic Tac Toe")
        self.root.geometry("600x600")
        self.root.resizable(True, True)

        self.board_frame = tk.Frame(self.root)
        self.board_frame.grid(row=0, column=0, sticky="nsew", padx=20, pady=(20,0))  # Place at the top and fill the available space

        self.info_label = tk.Label(self.root, text="Not in game lobby...")
        self.info_label.grid(row=1, column=0, ipadx=0, ipady=0, pady=10)  # Place at the center
        self.info_label.config(anchor="center")

        self.button_frame = tk.Frame(self.root)
        self.button_frame.grid(row=2, column=0, sticky="nsew", padx=20, pady=0)

        self.play_button = tk.Button(self.button_frame, text="PLAY")
        self.play_button.pack(side=tk.LEFT, fill=tk.BOTH, expand=True, padx=(0, 5))  # Add 5 pixels of padding on the right
        self.play_button.bind("<Button-1>", self.game_join)


        lower_frame = tk.Frame(self.root)
        lower_frame.grid(row=3, column=0, sticky="ew")  # Place at the bottom and fill horizontally

        self.status_label = tk.Label(self.root, text="")
        self.status_label.grid(row=4, column=0, ipadx=0, ipady=0, pady=10)  # Place at the center
        self.status_label.config(anchor="center")

        self.player1_label = tk.Label(lower_frame, text="You: " + client_status.name)
        self.player1_label.pack(side=tk.LEFT, padx=20)

        self.player2_label = tk.Label(lower_frame, text="Player 2")
        self.player2_label.pack(side=tk.RIGHT, padx=20)

        self.back_button = tk.Button(self.button_frame, text="BACK")
        self.back_button.pack(side=tk.RIGHT, fill=tk.BOTH, expand=True, padx=(5, 0))
        self.back_button.bind("<Button-1>", self.game_returntostart)
        self.back_button.pack_forget()

        self.reconnect_button = tk.Button(self.button_frame, text="RECONNECT")
        self.reconnect_button.pack(side=tk.RIGHT, fill=tk.BOTH, expand=True, padx=(5, 0))
        self.reconnect_button.bind("<Button-1>", self.send_recovery_manual)
        self.reconnect_button.pack_forget()

        self.render_board()
        self.automatic_reconnect_failed = False

        self.root.grid_rowconfigure(0, weight=1)  # Make the first row expandable
        self.root.grid_columnconfigure(0, weight=1)  # Make the first column expandable

        self.disable_buttons()  # Disable the buttons initially¨
        if self.client_status.recovery:
            self.tcp_client.send_data("", MsgRecoveryOpcode)

    def get_root(self):
        return self.root
    
    def get_pinger(self):
        return self.pinger
    
    def recovery_process(self, recovery_message, recovery_message_args):
        with self.recovery_mutex:
            if recovery_message == ClientMsgRecovery_InLobby:
                self.return_to_lobby_gui()
            elif recovery_message == ClientMsgRecovery_ReadyForGame:
                self.waiting_for_other_player_gui()
            elif recovery_message == ClientMsgRecovery_InGame_YourTurn:
                self.render_data_gui(recovery_message_args[0])
                self.set_other_player_name_gui(recovery_message_args[1])
                self.your_turn_gui()
            elif recovery_message == ClientMsgRecovery_InGame_OtherTurn:
                self.render_data_gui(recovery_message_args[0])
                self.set_other_player_name_gui(recovery_message_args[1])
                self.other_player_turn_gui()
            elif recovery_message == ClientMsgRecovery_InGame_OtherPlayAgain:
                self.render_data_gui(recovery_message_args[0])
                self.set_other_player_name_gui(recovery_message_args[1])
                self.waiting_for_other_player_gui()
            elif recovery_message == ClientMsgRecovery_InGame_GameOver:
                self.render_data_gui(recovery_message_args[0])
                self.gameover_gui(recovery_message_args[1])
                self.set_other_player_name_gui(recovery_message_args[2])

            self.reconnect_button.pack_forget()
            self.recovery_finished = True

        
    def server_is_offline_gui(self):
        self.return_to_lobby_gui()
        self.info_label.config(text="Connection is offline... Trying to reconnect")
        self.info_label.update()
        self.play_button.config(state=tk.DISABLED)
        self.disable_buttons()
        
    def send_recovery(self, interval: int = PingTime, amount: int = RecoveryMaxAttempts):
        self.server_is_offline_gui()
        self.reconnect_button.pack_forget()
        for i in range(amount):
            time.sleep(interval)
            print(log_message("Trying to automatically recover connection. Attempt " + str(i+1) + " of " + str(amount) + "..."))
            if self.try_recover():
                self.reconnect_button.pack_forget()
                self.automatic_reconnect_failed = False
                return
        self.reconnect_button.pack()
        self.info_label.config(text="Automatic reconnect failed. Click RECONNECT to try reconnect manually.")
        self.info_label.update()
        self.automatic_reconnect_failed = True

    def try_recover(self) -> bool:
        with self.recovery_mutex:
            try:
                self.tcp_client.reconnect()    
            except Exception as e:
                print(log_message("Attempt failed: " + str(e)))
                return False

            self.tcp_client.send_data(self.client_status.name, MsgLoginOpcode)
            self.tcp_client.send_data("", MsgRecoveryOpcode)
            if self.recovery_finished:
                self.recovery_finished = False
                self.restart_pinger()
                return True
            return False
        
    def send_recovery_manual(self, event):
        if self.automatic_reconnect_failed:
            print(log_message("Trying to manually recover connection..."))
            self.restart_pinger()
            self.reconnect_button.pack_forget()
    
    def restart_pinger(self):
        self.pinger.set_connection_online(True)
        self.pinger.run_pinger()

    def receive_from_dispatcher(self, opcode, reply_status, reply_msg: List[str]):
        reply_message = reply_msg[0]
        reply_message_extra = ""
        if len(reply_msg) > 1:
            reply_message_extra = reply_msg[1:]

        if opcode == MsgJoinOpcode:
            if reply_status == ClientMsgGameGone:
                self.return_to_lobby_gui()
            elif reply_status == ClientMsgOk:
                self.waiting_for_other_player_gui()
            else:
                print("error: " + reply_message)
        elif opcode == MsgMoveOpcode:
            if reply_status == ClientMsgOk:
                self.render_data_gui(reply_message)
                self.other_player_turn_gui()
            else:
                print("error: " + reply_message)
        elif opcode == MsgGameOverOpcode:
            if reply_status == ClientMsgOk:
                self.gameover_gui(reply_message)
            else:
                print("error: " + reply_message)
        elif opcode == MsgYourTurnOpcode:
            if reply_status == ClientMsgOk:
                self.your_turn_gui()
            else:
                print("error: " + reply_message)
        elif opcode == MsgGameStartedOpcode:
            if reply_status == ClientMsgOk:
                self.game_start_gui(reply_message)
            else:
                print("error: " + reply_message)
        elif opcode == MsgPlayAgainOpcode:
            if reply_status == ClientMsgOk:
                self.waiting_for_other_player_gui()
            elif reply_status == ClientMsgErr and reply_message == ClientMsgGameGone:
                self.set_status_label_gui(log_message("Game doesn't exist anymore (other player left?), returning to lobby"))
                self.return_to_lobby_gui()
            else:
                print("error: " + reply_message)
        elif opcode == MsgReturnToStartOpcode:
            if reply_status == ClientMsgOk:
                self.return_to_lobby_gui()
            elif reply_status == ClientMsgErr and reply_message == ClientMsgGameGone:
                #self.set_status_label_gui(self.log_message("Returned to lobby."))
                self.return_to_lobby_gui()
            else:
                print("error: " + reply_message)
        elif opcode == MsgRecoveryOpcode:
            if reply_status == ClientMsgOk:
                self.recovery_process(reply_message, reply_message_extra)
            else:
                print("error: " + reply_message)
        elif opcode == MsgPauseOpcode:
            if reply_status == ClientMsgOk:
                #text in form of [datetime] text
                self.set_status_label_gui(log_message("Other player disconnected, waiting for reconnection"))
            else:
                print("error: " + reply_message)
        elif opcode == MsgContinueOpcode:
            if reply_status == ClientMsgOk:
                self.set_status_label_gui(log_message("Other player reconnected, can continue"))
            else:
                print("error: " + reply_message)
        elif opcode == MsgStatusOpcode:
            if reply_status == ClientMsgOk:
                self.set_status_label_gui(log_message("Opponent has lost connection."))
            else:
                print("error: " + reply_message)
            

    def game_move(self, event, x, y):
        # Send the x, y position of the Tic Tac Toe move
        if event.widget["state"] == tk.DISABLED:
            return
        if self.tcp_client:
            print(log_message(f"Sending move: ({x}, {y})"))
            message = f"{x};{y}"
            self.tcp_client.send_data(message, MsgMoveOpcode)

    def game_join(self, event):
        if event.widget["state"] == tk.DISABLED:
            return
        print(log_message("Joining game..."))
        if self.tcp_client:
            message = ""
            self.tcp_client.send_data(message, MsgJoinOpcode)
                   
    def game_playagain(self, event):
        if event.widget["state"] == tk.DISABLED:
            return
        print(log_message("Requesting play again..."))
        if self.tcp_client:
            message = ""
            self.tcp_client.send_data(message, MsgPlayAgainOpcode)
            
    def game_returntostart(self, event):
        if event.widget["state"] == tk.DISABLED:
            return
        print(log_message("Returning to main menu..."))
        if self.tcp_client:
            message = ""
            self.tcp_client.send_data(message, MsgReturnToStartOpcode)

    def gameover_gui(self, winner_name: str):
        self.info_label.config(text=f"Game over! Winner: {winner_name}")
        self.info_label.update()
        self.disable_buttons()
        self.play_button.config(state=tk.NORMAL) 
        self.play_button.bind("<Button-1>", self.game_playagain)
        self.play_button.config(text="PLAY AGAIN")
        self.back_button.pack()

    def render_data_gui(self, data):
        data = data.split(rowSep)
        for row in range(self.client_status.board_size):
            for col in range(self.client_status.board_size):
                value = data[row].split(colSep)[col]
                if value == "1":
                    self.board_frame.grid_slaves(row=row, column=col)[0].config(text="X", state=tk.DISABLED, fg="red", font=("Helvetica", 24))
                elif value == "2":
                    self.board_frame.grid_slaves(row=row, column=col)[0].config(text="O", state=tk.DISABLED, fg="blue", font=("Helvetica", 24))
                else:
                    self.board_frame.grid_slaves(row=row, column=col)[0].config(text=" ", state=tk.DISABLED)

    def set_status_label_gui(self, text: str):
        self.status_label.config(text=text)
        self.status_label.update()

    def reset_status_label_gui(self):
        self.status_label.config(text="")
        self.status_label.update()

    def waiting_for_other_player_gui(self):
        self.info_label.config(text="Waiting for other player...")
        self.info_label.update()
        self.play_button.config(state=tk.DISABLED)
        self.disable_buttons()
        self.back_button.pack_forget()


    def return_to_lobby_gui(self):
        self.info_label.config(text="Not in game lobby...")
        self.info_label.update()
        self.play_button.config(state=tk.NORMAL)
        self.render_board()
        self.play_button.bind("<Button-1>", self.game_join)
        self.play_button.config(text="PLAY")
        self.back_button.pack_forget()
        self.set_other_player_name_gui("Player 2")
        self.disable_buttons()

    def set_other_player_name_gui(self, other_player_name: str):
        self.player2_label.config(text="Opponent: " + other_player_name)
        self.player2_label.update()

    def game_start_gui(self, other_player_name: str):
        self.other_player_turn_gui()
        self.set_other_player_name_gui(other_player_name)
        self.render_board()
        self.play_button.config(state=tk.DISABLED)
        self.enable_buttons()
        #self.status_label.config(text="OK")

    def your_turn_gui(self):
        self.info_label.config(text="Your turn")
        self.info_label.update()
        self.enable_buttons()

    def other_player_turn_gui(self):
        self.info_label.config(text="Other player's turn")
        self.info_label.update()
        self.disable_buttons()

    def render_board(self):
        for row in range(self.client_status.board_size):
            for col in range(self.client_status.board_size):
                button = tk.Button(self.board_frame, text=" ", width=4, height=2, font=("Helvetica", 24))
                button.grid(row=row, column=col, sticky="nsew")  # Add sticky="nsew" to make the button expand with the cell
                self.board_frame.grid_columnconfigure(col, weight=1)  # Make each column expandable
                self.board_frame.grid_rowconfigure(row, weight=1)  # Make each row expandable

                # Bind the button click event to a function that sends the x, y position
                button.bind("<Button-1>", lambda event, x=row, y=col: self.game_move(event, x, y))
                button.row, button.col = row, col

    def enable_buttons(self):
        # Enable all the buttons in the board
        for child in self.board_frame.winfo_children():
            child.config(state=tk.NORMAL)

    def disable_buttons(self):
        # Disable all the buttons in the board
        for child in self.board_frame.winfo_children():
            child.config(state=tk.DISABLED)

    def run(self):
        self.root.mainloop()


