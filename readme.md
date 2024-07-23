# Multiplayer Tic-Tac-Toe Game

This project is a multiplayer Tic-Tac-Toe game developed in Go and Python, designed to run on a server-client architecture. The server handles game logic, player connections, and communication, while the client provides a graphical user interface for players to interact with the game.

## Project Structure

- `client/`: Contains Python files for the client-side application.
  - `client_status.py`: Manages the client's status.
  - `connection.py`: Handles the connection to the server.
  - `const.py`: Defines constants used across the client application.
  - `dispatcher.py`: Dispatches messages to the appropriate handlers.
  - `game_gui.py`: Provides the graphical user interface for the game.
  - `IDispatchReceiver.py`: Interface for message dispatch receivers.
  - `login_gui.py`: Manages the login graphical user interface.
  - `login_processor.py`: Processes login requests.
  - `main.py`: The entry point for the client application.
  - `message_formatter.py`: Formats messages for sending to the server.
  - `pinger.py`: Sends periodic pings to the server to maintain the connection.
- `util/`: Contains Go files for utility functions and game logic.
  - `const.go`: Defines constants used across the server application.
  - `game.go`: Contains the game logic for Tic-Tac-Toe.
  - `player.go`: Manages player information and actions.
  - `server.go`: Handles server operations, including client connections and message routing.
- `go.mod`: Defines the Go module and its dependencies.
- `main.go`: The entry point for the server application.
- `readme.md`: This file, providing an overview of the project.
- 'KIV_UPS_SP.pdf': Documentation of this project written in Czech language.

## Getting Started

To run this project, you will need Go (1.15.15) and Python with tkinter installed on your system.

### Running the Server

1. Navigate to the project root directory.
2. Run `go1.15.15 run .` to run the server application.

### Running the Client

1. Navigate to the `client/` directory.
3. Run `python main.py` to start the client application.

