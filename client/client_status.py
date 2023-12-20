from const import ConnHost, ConnPort


class ClientStatus:
    """
    Represents the status (info) of a client.

    Attributes:
        name (str): The name of the client.
        board_size (int): The size of the board.
        logged_in (bool): Indicates whether the client is logged in.
        recovery (bool): Indicates whether the client is in recovery mode.
        server_ip (str): The IP address of the server.
        server_port (int): The port number of the server.
    """

    def __init__(self, name: str = "", board_size: int = 0, logged_in: bool = False, recovery: bool = False, server_ip: str = ConnHost, server_port: int = ConnPort):
        self.name = name
        self.board_size = board_size
        self.logged_in = logged_in
        self.server_ip = server_ip
        self.server_port = server_port
        self.recovery = recovery