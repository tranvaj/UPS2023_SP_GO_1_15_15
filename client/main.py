import threading
import tkinter as tk
import tkinter.messagebox
from client_status import ClientStatus
from login_gui import LoginScreen
from login_processor import LoginProcessor
from connection import TCPClient
from const import ConnHost, ConnPort
from game_gui import TicTacToeGUI
from dispatcher import Dispatcher
from pinger import Pinger

tcp_client = TCPClient(ConnHost, ConnPort)

try:
    client_status = ClientStatus()
    login_screen = LoginScreen(client_status=client_status)
    login_screen.mainloop()
    print("Login screen closed.")

    try:
        login_processor = LoginProcessor(tcp_client=tcp_client, client_status=client_status)
    except Exception as e:
        print("Shutting down due to Error: " + str(e))
        exit()
        
    dispatcher = Dispatcher(tcp_client, login=login_processor, game_gui=None, pinger=None)
    dispatcher.run_dispatcher()
    login_processor.wait_for_data()

    if client_status.logged_in == False:
        print("Client not logged in. Closing.")
        exit()
        
    game_gui = TicTacToeGUI(tcp_client=tcp_client, client_status=client_status)
    pinger = game_gui.get_pinger()
    dispatcher.set_pinger(pinger)
    dispatcher.set_game_gui(game_gui)
    pinger.run_pinger()
    game_gui.run()
except Exception as e:
    print("error: " + str(e))
    exit()


