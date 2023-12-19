from const import ConnHost, ConnPort


class ClientStatus:
    def __init__(self, name: str = "", board_size: int = 0, logged_in: bool = False, recovery: bool = False, server_ip: str = ConnHost, server_port: int = ConnPort):
        self.name = name
        self.board_size = board_size
        self.logged_in = logged_in
        self.server_ip = server_ip
        self.server_port = server_port
        self.recovery = recovery