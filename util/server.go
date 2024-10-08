package util

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var availableGamesList = make([]*TicTacToeGame, 0) //list of available games
var gameListMutex = &sync.Mutex{}                  //mutex for availableGamesList (thread safety)
var players = NewPlayers()                         //list of players

// readAll reads data from the connection until the specified data length is reached.
// It returns the read data, the total number of bytes read, and any error encountered.
// If the data length exceeds the maximum allowed size, it returns an error.
// If a timeout is specified, it sets a read deadline on the connection.
// The read deadline is cleared before returning.
func readAll(connection *net.Conn, dataLen int, timeout int) ([]byte, int, error) {
	totalBytesRead := 0
	conn := *connection
	buffer := make([]byte, dataLen)

	if dataLen > MaxDataLen {
		return nil, 0, fmt.Errorf("data size is too large")
	}
	if timeout != 0 {
		conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(timeout)))
	}

	for totalBytesRead < dataLen {
		bytesRead, err := conn.Read(buffer[totalBytesRead:])
		//println("yessir", bytesRead)
		if err != nil {
			return nil, totalBytesRead, err
		}
		totalBytesRead += bytesRead
	}
	defer conn.SetReadDeadline(time.Time{})
	return buffer, totalBytesRead, nil
}

// writeAll writes the given data to the provided connection until all bytes are written or an error occurs.
// It sets a write deadline if a timeout is specified.
// The function returns the total number of bytes written and any error encountered.
func writeAll(connection *net.Conn, data []byte, timeout int) (int, error) {
	totalBytesWritten := 0
	conn := *connection

	if timeout != 0 {
		conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(timeout)))
	}

	for totalBytesWritten < len(data) {
		bytesWritten, err := conn.Write(data[totalBytesWritten:])
		if err != nil {
			return bytesWritten, err
		}
		totalBytesWritten += bytesWritten
	}
	defer conn.SetWriteDeadline(time.Time{})
	return totalBytesWritten, nil
}

// ProcessClient handles the communication with a client.
// It reads messages from the client, processes the requested operation,
// and sends back the response. If the client sends too many invalid operations,
// the connection is closed.
//
// Parameters:
// - connection: The network connection with the client.
// - player: A pointer to the Player struct representing the client.
//
// Note: This function should be called as a goroutine to handle multiple clients concurrently.
func ProcessClient(connection net.Conn, player *Player) {
	defer connection.Close()
	invalidOp := 0
	for {
		msg, _, err := readAll(&connection, MsgHeaderLen, 0)
		if err != nil {
			fmt.Println("could not read client message, closing", err)
			return
		}

		msgHeader := string(msg[0:len(MsgMagic)])
		opcode := string(msg[len(MsgMagic) : len(MsgMagic)+len(MsgLoginOpcode)])
		if msgHeader != MsgMagic {
			fmt.Println("msg header was incorrect")
			return
		}

		dataLen, err := strconv.Atoi(string(msg[len(MsgMagic)+len(MsgLoginOpcode):]))
		if err != nil {
			fmt.Println("couldnt get data length", string(msg))
			continue
		}

		//wait for data
		//fmt.Println("waiting for bytes:", dataLen, string(msg))
		data, _, err := readAll(&connection, dataLen, 0)
		if err != nil {
			fmt.Println(err)
			return
		}

		log.Println(fmt.Sprintf("Received from %s message: %s", connection.RemoteAddr().String(), string(msg)+string(data)))

		if player.Conn == nil && opcode != MsgLoginOpcode && opcode != MsgPingOpcode {
			_, err := sendMsg(&connection, createOpCode(opcode, false, "Only logged in clients can execute commands other than ping."), 0)
			if err != nil {
				fmt.Println("could not send message to client")
				return
			}
		} else {
			opMessage, err := processOperation(&player, &connection, opcode, strings.Split(string(data), ArgSep))
			messageToSend := opMessage
			success := true //represenets status of operation
			if err != nil {
				fmt.Println("could not process operation", err)
				messageToSend = err.Error()
				success = false
			} else if opcode == MsgLoginOpcode {
				messageToSend = opMessage + ArgSep + fmt.Sprint(defaultBoardSize) //send board size
			}

			if !(opMessage == "" && err == nil) { //if string is empty and err is nil means "dont send response" (its handled in processOperation)
				messageToSend = createOpCode(opcode, success, messageToSend)
				msgArg := strings.Split(messageToSend, ArgSep)
				msgLastArg := msgArg[len(msgArg)-1]
				if msgLastArg == SrvErrInvalidOp {
					invalidOp++
					if invalidOp >= MaxInvalidOp {
						log.Printf(fmt.Sprintf("Client %s sent too many invalid operations (%d), closing connection\n", connection.RemoteAddr().String(), invalidOp))
						playerDisconnected(player)
						return
					}
				}
				_, err = sendMsg(&connection, messageToSend, 0)
				//log.Default().Println("Sent to:", connection.RemoteAddr().String(), messageToSend)
				if err != nil {
					fmt.Println("could not send message to client")
					//return
				}
			}
		}
		//fmt.Printf("%v\n", player)
	}
}

// removeGame removes a game from the available games list based on the given gameId.
// It acquires a lock on the gameListMutex to ensure thread safety.
// The game is removed by slicing the availableGamesList and reassigning it.
func removeGame(gameId int) {
	gameListMutex.Lock()
	defer gameListMutex.Unlock()
	if gameId < 0 {
		log.Println("Game doesn't exist")
		return
	}
	availableGamesList = append(availableGamesList[:gameId], availableGamesList[gameId+1:]...)
}

// playerDisconnected handles the disconnection of a player.
//
// It logs out the player, finds the game the player was in, and performs necessary actions based on the game state and the other player's status.
// If the other player is ready for a game and the game is over, it sends a message to return to lobby (where the player can find another player to play with).
// If the other player is in a game and the game is not over, it sends a message to the other player indicating that the opponent has disconnected.
// It also sends a message to the other player indicating that the opponent has lost connection.
//
// Finally, it removes the player from the game and removes the game if necessary.
func playerDisconnected(player *Player) {
	players.Logout(player)
	game := findGame(player)
	if game != nil {
		otherPlayer := game.GetOtherPlayer(player)
		game.RemovePlayer(player)
		if otherPlayer.Id != 0 {
			if otherPlayer.Status == ReadyForGame && game.gameState == GameOver {
				otherPlayer.Status = InLobby
				_, err := sendMsg(otherPlayer.Conn, createOpCode(MsgPlayAgainOpcode, false, ClientMsgGameGone), 0)
				if err != nil {
					log.Println("could not send return to start to player two")
				}
			} else if otherPlayer.Status == InGame && game.gameState != GameOver {
				_, err := sendMsg(otherPlayer.Conn, createOpCode(MsgGameOverOpcode, true, otherPlayer.Name+"(Opponent disconnected)"), 0)
				if err != nil {
					log.Println("could not send game over to player two")
				}
			}
			_, err := sendMsg(otherPlayer.Conn, createOpCode(MsgStatusOpcode, true, "Opponent has lost connection."), 0)
			if err != nil {
				log.Println("could not send status to other player")
			}
		}
		removeGame(getGameId(game))
	}
}

// broadcastMsg sends the given message to all connections in the given slice.
func broadcastMsg(connections []*net.Conn, msg string, timeout int) []error {
	log.Println(fmt.Sprintf("Broadcasting message to %d clients: %s", len(connections), msg))
	errs := make([]error, 0)
	for _, conn := range connections {
		_, err := sendMsg(conn, msg, timeout)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// sendMsg sends the given message to the given connection.
func sendMsg(connection *net.Conn, msg string, timeout int) (int, error) {
	bytesWritten, err := writeAll(connection, []byte(msg), timeout)
	log.Println(fmt.Sprintf("Sent to %s message: %s", (*connection).RemoteAddr().String(), msg))
	return bytesWritten, err
}

// createOpCode creates an protocol response with the given parameters.
func createOpCode(opcode string, success bool, data string) string {
	if success {
		data = ClientMsgOk + ArgSep + data
	} else {
		data = ClientMsgErr + ArgSep + data
	}
	return MsgMagic + opcode + fmt.Sprintf("%04d", len(data)) + data
}

// processOperation processes the operation based on the given opcode and data.
// It updates the player's status and game state accordingly.
// Handles recovery of player state in client.
// If an error occurs during the operation, it returns an error message.
// Otherwise, it returns a success message or an empty string.
func processOperation(playerAddress **Player, conn *net.Conn, opcode string, data []string) (string, error) {
	player := *playerAddress
	var err error = nil
	var game *TicTacToeGame
	if player.Id == 0 {
		game = nil
	} else {
		game = findGame(player)

		if !player.Connected && opcode != MsgRecoveryOpcode {
			return "", fmt.Errorf("must send recovery opcode after reconnection")
		}
		if game != nil && player.Connected && opcode != MsgPingOpcode {
			otherPlayer := game.GetOtherPlayer(player)
			if !otherPlayer.Connected && otherPlayer.Id != 0 {
				informPlayerAboutDisconnect(player)
				return "", fmt.Errorf("other player disconnected, must wait for other player") //s
			}
		}
	}

	switch opcode {
	case MsgLoginOpcode:
		relogin := false
		if len(data) != 1 {
			return "", fmt.Errorf("wrong number of arguments")
		}
		if len(data[0]) == 0 {
			return "", fmt.Errorf("name cannot be empty")
		}
		loginPlayer, err := players.Login(conn, data[0], player) //if no err -> replace old player with new one
		if err != nil {
			//didnt find player
			//add
			player.Name = data[0]
			err := players.AddNewPlayer(player)
			if err != nil {
				return "", fmt.Errorf(err.Error())
			}
		} else {
			*playerAddress = loginPlayer
			player = *playerAddress
			relogin = true
		}

		updatePlayerConnected(player)
		if relogin {
			player.Connected = false //go call recovery msg
			return "", fmt.Errorf(ClientMsgRecoveryLogin + ArgSep + fmt.Sprint(defaultBoardSize))
		} else {
			go disconnectHandler(player)
			go ConnectionCloseHandler(player)
			return fmt.Sprintf("Welcome %s. Your ID is: %d", player.Name, player.Id), nil
		}
	case MsgJoinOpcode:
		if player.Status != InLobby {
			return "", fmt.Errorf("player not in lobby" + ArgSep + SrvErrInvalidOp)
		}
		game := operationJoin(player)

		err = game.Start()
		if err != nil {
			log.Println(err.Error())
			player.Status = ReadyForGame
			return fmt.Sprintf("joined game %d", getGameId(game)), nil
		}
		otherPlayer := game.GetOtherPlayer(player)
		//broadcast game started
		_, err := sendMsg(player.Conn, createOpCode(MsgGameStartedOpcode, true, otherPlayer.Name), 0)
		if err != nil {
			log.Println("could not send game started to player one")
		}
		_, err = sendMsg(otherPlayer.Conn, createOpCode(MsgGameStartedOpcode, true, player.Name), 0)
		if err != nil {
			log.Println("could not send game started to player two")
		}

		game.players[0].Status = InGame
		game.players[1].Status = InGame
		//tell player one to move
		_, err = sendMsg(game.players[0].Conn, createOpCode(MsgYourTurnOpcode, true, ""), 0)
		if err != nil {
			log.Println("could not send move to player one")
		}
		return "", nil

	case MsgMoveOpcode:
		if len(data) != 2 {
			return "", fmt.Errorf("wrong number of arguments" + ArgSep + SrvErrInvalidOp)
		}
		if game == nil || player.Status != InGame {
			return "", fmt.Errorf("player not in game" + ArgSep + SrvErrInvalidOp)
		}
		if game.gameState != WaitingForPlayerOneMove && game.gameState != WaitingForPlayerTwoMove {
			return "", fmt.Errorf("game not in play state" + ArgSep + SrvErrInvalidOp)
		}
		otherPlayer := game.GetOtherPlayer(player)
		if !otherPlayer.Connected && otherPlayer.Id != 0 {
			informPlayerAboutDisconnect(player)
			return "", fmt.Errorf("move: other player disconnected, must wait for other player")
		}

		x, err := strconv.Atoi(data[0])
		if err != nil {
			return "", fmt.Errorf("couldnt parse arg" + ArgSep + SrvErrInvalidOp)
		}
		y, err := strconv.Atoi(data[1])
		if err != nil {
			return "", fmt.Errorf("couldnt parse arg" + ArgSep + SrvErrInvalidOp)
		}
		if game == nil {
			return "", fmt.Errorf("player not in game" + ArgSep + SrvErrInvalidOp)
		}
		err = game.Move(*player, x, y)
		if err != nil {
			return "", fmt.Errorf(err.Error() + ArgSep + SrvErrInvalidOp)
		}

		//broadcast board in string format
		board := game.GetBoardInParsableFormat()
		errs := broadcastMsg([]*net.Conn{game.players[0].Conn, game.players[1].Conn}, createOpCode(MsgMoveOpcode, true, board), 0)
		if errs != nil {
			log.Println("could not broadcast board to all players")
		}

		if game.gameOverState != NotOver {
			//game is over
			result := ""
			winner := game.GetGameWinner()
			if winner == nil {
				result = "Draw"
			} else {
				result = winner.Name
			}
			errs := broadcastMsg([]*net.Conn{game.players[0].Conn, game.players[1].Conn}, createOpCode(MsgGameOverOpcode, true, result), 0)
			if errs != nil {
				log.Println("could not broadcast game over to all players")
			}
			//game.Reset(true)
			return "", nil
		}

		//tell other player to move
		if game.gameState == WaitingForPlayerOneMove {
			_, err = sendMsg(game.players[0].Conn, createOpCode(MsgYourTurnOpcode, true, ""), 0)
			if err != nil {
				log.Println("could not send move to player two")
			}
		} else if game.gameState == WaitingForPlayerTwoMove {
			_, err = sendMsg(game.players[1].Conn, createOpCode(MsgYourTurnOpcode, true, ""), 0)
			if err != nil {
				log.Println("could not send move to player one")
			}
		}
		return "", nil

	case MsgPlayAgainOpcode:
		if game == nil {
			player.Status = InLobby
			return "", fmt.Errorf(ClientMsgGameGone)
		}
		if !(player.Status == InGame && game.gameState == GameOver) {
			return "", fmt.Errorf("player not in game or game not over" + ArgSep + SrvErrInvalidOp)
		}

		err = game.PlayAgain(*player)
		if err != nil {
			return "", fmt.Errorf(err.Error() + ArgSep + SrvErrInvalidOp)
		}
		player.Status = ReadyForGame

		err = game.Start()
		if err != nil {
			log.Println(err.Error())
			return fmt.Sprintf("requesting play again (game id: %d)", getGameId(game)), nil
		}

		otherPlayer := game.GetOtherPlayer(player)
		//send game started with opponent name
		_, err := sendMsg(player.Conn, createOpCode(MsgGameStartedOpcode, true, otherPlayer.Name), 0)
		if err != nil {
			log.Println("could not send game started to player one")
		}
		_, err = sendMsg(otherPlayer.Conn, createOpCode(MsgGameStartedOpcode, true, player.Name), 0)
		if err != nil {
			log.Println("could not send game started to player two")
		}
		game.players[0].Status = InGame
		game.players[1].Status = InGame

		//change starting player
		if game.gameState == WaitingForPlayerOneMove {
			game.gameState = WaitingForPlayerTwoMove
		} else {
			game.gameState = WaitingForPlayerOneMove
		}
		//tell player two to move
		_, err = sendMsg(game.players[1].Conn, createOpCode(MsgYourTurnOpcode, true, ""), 0)
		if err != nil {
			log.Println("could not send move to player two")
		}
		return "", nil
	case MsgReturnToStartOpcode:
		if game == nil {
			player.Status = InLobby
			return "", fmt.Errorf(ClientMsgGameGone)
		}
		if !(player.Status == InGame && game.gameState == GameOver) {
			return "", fmt.Errorf("player not in game or game not over" + ArgSep + SrvErrInvalidOp)
		}
		player.Status = InLobby
		otherPlayer := game.GetOtherPlayer(player)

		if otherPlayer.Status == ReadyForGame {
			otherPlayer.Status = InLobby
			_, err = sendMsg(otherPlayer.Conn, createOpCode(MsgPlayAgainOpcode, false, ClientMsgGameGone), 0)
			if err != nil {
				log.Println("could not send return to start to player two")
			}
		}
		game.Reset(false)
		removeGame(getGameId(game)) //player left, removing game
		return "left the lobby", nil
	case MsgPingOpcode:
		player.TimeSinceLastPing = time.Now()
		return "ping", nil
	case MsgRecoveryOpcode:
		return handleRecoveryOpcode(player, game)
	default:
		return "", fmt.Errorf("unknown opcode")
	}
}

// handleRecoveryOpcode handles the recovery operation code for a player in a TicTacToe game.
// It takes a player pointer and a game pointer as parameters and returns a string and an error.
// The string represents the recovery option for the player, while the error indicates any error that occurred during the operation.
func handleRecoveryOpcode(player *Player, game *TicTacToeGame) (string, error) {
	option := ""
	var err error
	if player.Status == InLobby || (game == nil && player.Status == InGame) {
		player.Status = InLobby //game gone
		option = ClientMsgRecovery_InLobby
	} else if player.Status == ReadyForGame {
		option = ClientMsgRecovery_ReadyForGame
	} else if player.Status == InGame {
		result := ""
		winner := game.GetGameWinner()
		if winner == nil {
			result = "Draw"
		} else {
			result = winner.Name
		}
		otherPlayer := game.GetOtherPlayer(player)
		otherPlayerName := ""
		if otherPlayer.Id != 0 {
			otherPlayerName = otherPlayer.Name
		}
		board := game.GetBoardInParsableFormat()
		if game.gameState == WaitingForPlayerOneMove && player.Id == game.players[0].Id {
			option = ClientMsgRecovery_InGame_YourTurn + ArgSep + board + ArgSep + otherPlayerName
		} else if game.gameState == WaitingForPlayerOneMove && player.Id == game.players[1].Id {
			option = ClientMsgRecovery_InGame_OtherTurn + ArgSep + board + ArgSep + otherPlayerName
		} else if game.gameState == WaitingForPlayerTwoMove && player.Id == game.players[1].Id {
			option = ClientMsgRecovery_InGame_YourTurn + ArgSep + board + ArgSep + otherPlayerName
		} else if game.gameState == WaitingForPlayerTwoMove && player.Id == game.players[0].Id {
			option = ClientMsgRecovery_InGame_OtherTurn + ArgSep + board + ArgSep + otherPlayerName
		} else if game.gameState == GameOver {
			option = ClientMsgRecovery_InGame_GameOver + ArgSep + board + ArgSep + result + ArgSep + otherPlayerName
		}
	} else {
		return "", fmt.Errorf("unknown player state")
	}
	if !player.Connected {
		player.Connected = true
		if game != nil {
			otherPlayer := game.GetOtherPlayer(player)
			if otherPlayer.Id != 0 {
				_, err = sendMsg(otherPlayer.Conn, createOpCode(MsgContinueOpcode, true, ""), 0)
				if err != nil {
					log.Println("could not send continue to other player")
				}
			}
		}
		go disconnectHandler(player)
	}
	player.TimeSinceLastPing = time.Now()
	return option, nil
}

// informPlayerAboutDisconnect sends a message to the given player indicating that the opponent has disconnected.
func informPlayerAboutDisconnect(player *Player) {
	game := findGame(player)
	if game == nil {
		return
	}
	otherPlayer := game.GetOtherPlayer(player)
	if otherPlayer.Id == 0 {
		return
	}
	_, err := sendMsg(player.Conn, createOpCode(MsgPauseOpcode, true, ""), 0)
	if err != nil {
		fmt.Println("could not send message to client")
		return
	}
}

// ConnectionCloseHandler handles the closing of a connection.
// Always one per player.
// Closes connection and removes the player has not pinged in a while (timeouted).
func ConnectionCloseHandler(player *Player) {
	log.Println("!!! Starting connection close handler for player " + player.Name + "!!!")
	for {
		time.Sleep(time.Second * PingTime)
		if player.getTimeSinceLastPing() > time.Second*MaxSecondsBeforeDisconnect {
			log.Println(fmt.Sprintf("Player %s (ID: %d) timed out, closing connection", player.Name, player.Id))

			playerDisconnected(player)
			if player.Conn == nil {
				return
			}
			playerConn := (*player.Conn)
			if playerConn != nil {
				err := playerConn.Close()
				if err != nil {
					log.Println("could not close connection")
				}
			}
			return
		}
	}
}

// Checks if player has disconnected
func disconnectHandler(player *Player) {
	playerCopy := *player
	for {
		time.Sleep(time.Second * PingTime)
		updatePlayerConnected(player)
		if !player.Connected || playerCopy.Conn != player.Conn {
			log.Printf("Player %s (ID: %d) disconnected\n", player.Name, player.Id)
			game := findGame(player)
			if game == nil {
				return
			}
			otherPlayer := game.GetOtherPlayer(player)
			if otherPlayer.Id == 0 {
				return
			}
			informPlayerAboutDisconnect(otherPlayer)
			return
		}
	}
}

// Sets player.Connected value based on PingTime and MaxNoPingReceived
func updatePlayerConnected(player *Player) {
	if player.Conn != nil {
		if player.getTimeSinceLastPing() > time.Second*PingTime*MaxNoPingReceived {
			player.Connected = false
		} else {
			player.Connected = true
		}
	} else {
		player.Connected = false
	}
}

// Get game id in list of available games
func getGameId(game *TicTacToeGame) int {
	gameListMutex.Lock()
	defer gameListMutex.Unlock()
	for i, v := range availableGamesList {
		if v == game {
			return i
		}
	}
	return -1
}

// Finds game that player is in or joins existing game that is not full or creates new game if none found
func operationJoin(player *Player) *TicTacToeGame {
	game := findGame(player)
	if game == nil {
		game = joinGame(player)
	}
	if game == nil {
		game = createGame()
		game.Join(player)
	}
	return game
}

// Find game that player is in
func findGame(player *Player) *TicTacToeGame {
	gameListMutex.Lock()
	defer gameListMutex.Unlock()
	for _, v := range availableGamesList {
		if v.players[0].Id == player.Id || v.players[1].Id == player.Id {
			return v
		}
	}
	//player not in any game
	return nil
}

// Join game that is not full
func joinGame(player *Player) *TicTacToeGame {
	gameListMutex.Lock()
	defer gameListMutex.Unlock()
	for i, v := range availableGamesList {
		if v.players[0].Id == 0 || v.players[1].Id == 0 {
			v.Join(player)
			return availableGamesList[i]
		}
	}
	//no available games
	return nil
}

// Create a new game
func createGame() *TicTacToeGame {
	newGame := NewTickTackToeGame(defaultBoardSize)
	gameListMutex.Lock()
	defer gameListMutex.Unlock()
	availableGamesList = append(availableGamesList, newGame)
	return newGame
}
