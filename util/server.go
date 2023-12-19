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

var availableGamesList = make([]*TicTacToeGame, 0)
var gameListMutex = &sync.Mutex{}
var players = NewPlayers()

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

func ProcessClient(connection net.Conn) {
	defer connection.Close()
	invalidOp := 0
	player := &Player{}
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
			continue
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

		fmt.Println("Received: ", connection.RemoteAddr().String(), string(msg), string(data))

		if player.conn == nil && opcode != MsgLoginOpcode {
			_, err := sendMsg(&connection, createOpCode(opcode, false, "Only logged in clients can execute commands."), 0)
			if err != nil {
				fmt.Println("could not send message to client")
				return
			}
		} else {
			opMessage, err := processOperation(&player, &connection, opcode, strings.Split(string(data), ArgSep))
			messageToSend := opMessage
			success := true
			if err != nil {
				fmt.Println("could not process operation", err)
				messageToSend = err.Error()
				success = false
			} else if opcode == MsgLoginOpcode {
				messageToSend = opMessage + ArgSep + fmt.Sprint(defaultBoardSize) //send board size
			}

			if !(opMessage == "" && err == nil) { //if string is empty and err is nil, dont send anything (its handled in processOperation)
				messageToSend = createOpCode(opcode, success, messageToSend)
				msgArg := strings.Split(messageToSend, ArgSep)
				msgLastArg := msgArg[len(msgArg)-1]
				if msgLastArg == SrvErrInvalidOp {
					invalidOp++
					if invalidOp >= MaxInvalidOp {
						log.Printf(fmt.Sprintf("Client %s sent too many invalid operations (%d), closing connection\n", connection.RemoteAddr().String(), invalidOp))
						return
					}
				}
				_, err = sendMsg(&connection, messageToSend, 0)
				//log.Default().Println("Sent to:", connection.RemoteAddr().String(), messageToSend)
				if err != nil {
					fmt.Println("could not send message to client")
					return
				}
			}

			if *player != (Player{}) && opcode == MsgLoginOpcode {
				defer playerDisconnected(player) //if player logged in, defer disconnect
			}
		}
		//fmt.Printf("%v\n", player)
	}
}

func removeGame(gameId int) {
	gameListMutex.Lock()
	defer gameListMutex.Unlock()
	availableGamesList = append(availableGamesList[:gameId], availableGamesList[gameId+1:]...)
}

func playerDisconnected(player *Player) {
	players.Logout(player)
	game := findGame(player)
	if game != nil {
		game.RemovePlayer(player)
		otherPlayer := game.GetOtherPlayer(player)
		if otherPlayer.id != 0 {
			if game.gameState != GameOver {
				//other player won
				_, err := sendMsg(otherPlayer.conn, createOpCode(MsgGameOverOpcode, true, otherPlayer.name+" (opponent disconnected)"), 0)
				if err != nil {
					fmt.Println("could not send game over to other player")
					return
				}
			}
		}
		removeGame(getGameId(game))
	}
}

func broadcastMsg(connections []*net.Conn, msg string, timeout int) []error {
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

func sendMsg(connection *net.Conn, msg string, timeout int) (int, error) {
	bytesWritten, err := writeAll(connection, []byte(msg), timeout)
	log.Println("Sent to:", (*connection).RemoteAddr().String(), msg)
	return bytesWritten, err
}

func createOpCode(opcode string, success bool, data string) string {
	if success {
		data = ClientMsgOk + ArgSep + data
	} else {
		data = ClientMsgErr + ArgSep + data
	}
	return MsgMagic + opcode + fmt.Sprintf("%04d", len(data)) + data
}

func processOperation(playerAddress **Player, conn *net.Conn, opcode string, data []string) (string, error) {
	player := *playerAddress
	var err error = nil
	var game *TicTacToeGame
	if player.id == 0 {
		game = nil
	} else {
		game = findGame(player)

		if !player.connected && opcode != MsgRecoveryOpcode {
			return "", fmt.Errorf("must send recovery opcode after reconnection")
		}
		if game != nil && player.connected && opcode != MsgPingOpcode {
			otherPlayer := game.GetOtherPlayer(player)
			if !otherPlayer.connected && otherPlayer.id != 0 {
				informPlayerAboutDisconnect(player)
				return "", fmt.Errorf("other player disconnected, must wait for other player")
			}
		}
	}

	switch opcode {
	case MsgLoginOpcode:
		relogin := false
		if players.GetPlayerIndex(data[0]) != -1 {
			relogin = true
		}
		if len(data) != 1 {
			return "", fmt.Errorf("wrong number of arguments")
		}
		if len(data[0]) == 0 {
			return "", fmt.Errorf("name cannot be empty")
		}
		loginPlayer, err := players.Login(conn, data[0])
		if err != nil {
			return "", err
		}
		*playerAddress = loginPlayer
		player = *playerAddress

		updatePlayerConnected(player)
		if relogin {
			return "", fmt.Errorf(ClientMsgRecoveryLogin + ArgSep + fmt.Sprint(defaultBoardSize))
		} else {
			go disconnectHandler(player)
			go connectionCloseHandler(player)
			return fmt.Sprintf("Welcome %s. Your ID is: %d", player.name, player.id), nil
		}
	case MsgJoinOpcode:
		if player.status != InLobby {
			return "", fmt.Errorf("player not in lobby" + ArgSep + SrvErrInvalidOp)
		}
		game := operationJoin(player)

		err = game.Start()
		if err != nil {
			log.Println(err.Error())
			player.status = ReadyForGame
			return fmt.Sprintf("joined game %d", getGameId(game)), nil
		}
		otherPlayer := game.GetOtherPlayer(player)
		//broadcast game started
		_, err := sendMsg(player.conn, createOpCode(MsgGameStartedOpcode, true, otherPlayer.name), 0)
		if err != nil {
			return "", fmt.Errorf("could not send game started to player one")
		}
		_, err = sendMsg(otherPlayer.conn, createOpCode(MsgGameStartedOpcode, true, player.name), 0)
		if err != nil {
			return "", fmt.Errorf("could not send game started to player two")
		}

		game.players[0].status = InGame
		game.players[1].status = InGame
		//tell player one to move
		_, err = sendMsg(game.players[0].conn, createOpCode(MsgYourTurnOpcode, true, ""), 0)
		if err != nil {
			return "", fmt.Errorf("could not send move to player one")
		}
		return "", nil

	case MsgMoveOpcode:
		if len(data) != 2 {
			return "", fmt.Errorf("wrong number of arguments" + ArgSep + SrvErrInvalidOp)
		}
		if game == nil || player.status != InGame {
			return "", fmt.Errorf("player not in game" + ArgSep + SrvErrInvalidOp)
		}
		if game.gameState != WaitingForPlayerOneMove && game.gameState != WaitingForPlayerTwoMove {
			return "", fmt.Errorf("game not in play state" + ArgSep + SrvErrInvalidOp)
		}
		otherPlayer := game.GetOtherPlayer(player)
		if !otherPlayer.connected && otherPlayer.id != 0 {
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
		errs := broadcastMsg([]*net.Conn{game.players[0].conn, game.players[1].conn}, createOpCode(MsgMoveOpcode, true, board), 0)
		if errs != nil {
			return "", fmt.Errorf("could not broadcast board to all players")
		}

		if game.gameOverState != NotOver {
			//game is over
			result := ""
			winner := game.GetGameWinner()
			if winner == nil {
				result = "Draw"
			} else {
				result = winner.name
			}
			errs := broadcastMsg([]*net.Conn{game.players[0].conn, game.players[1].conn}, createOpCode(MsgGameOverOpcode, true, result), 0)
			if errs != nil {
				return "", fmt.Errorf("could not broadcast game over to all players")
			}
			//game.Reset(true)
			return "", nil
		}

		//tell other player to move
		if game.gameState == WaitingForPlayerOneMove {
			_, err = sendMsg(game.players[0].conn, createOpCode(MsgYourTurnOpcode, true, ""), 0)
			if err != nil {
				return "", fmt.Errorf("could not send move to player two")
			}
		} else if game.gameState == WaitingForPlayerTwoMove {
			_, err = sendMsg(game.players[1].conn, createOpCode(MsgYourTurnOpcode, true, ""), 0)
			if err != nil {
				return "", fmt.Errorf("could not send move to player one")
			}
		}
		return "", nil

	case MsgPlayAgainOpcode:
		if game == nil {
			player.status = InLobby
			return "", fmt.Errorf(ClientMsgGameGone)
		}
		if !(player.status == InGame && game.gameState == GameOver) {
			return "", fmt.Errorf("player not in game or game not over" + ArgSep + SrvErrInvalidOp)
		}

		err = game.PlayAgain(*player)
		if err != nil {
			return "", fmt.Errorf(err.Error() + ArgSep + SrvErrInvalidOp)
		}
		player.status = ReadyForGame

		err = game.Start()
		if err != nil {
			log.Println(err.Error())
			return fmt.Sprintf("play again game %d", getGameId(game)), nil
		}
		//broadcast game started
		log.Println("broadcasting game started")
		otherPlayer := game.GetOtherPlayer(player)
		//broadcast game started
		_, err := sendMsg(player.conn, createOpCode(MsgGameStartedOpcode, true, otherPlayer.name), 0)
		if err != nil {
			return "", fmt.Errorf("could not send game started to player one")
		}
		_, err = sendMsg(otherPlayer.conn, createOpCode(MsgGameStartedOpcode, true, player.name), 0)
		if err != nil {
			return "", fmt.Errorf("could not send game started to player two")
		}
		game.players[0].status = InGame
		game.players[1].status = InGame

		//change starting player
		if game.gameState == WaitingForPlayerOneMove {
			game.gameState = WaitingForPlayerTwoMove
		} else {
			game.gameState = WaitingForPlayerOneMove
		}
		//tell player two to move
		_, err = sendMsg(game.players[1].conn, createOpCode(MsgYourTurnOpcode, true, ""), 0)
		if err != nil {
			return "", fmt.Errorf("could not send move to player two")
		}
		return "", nil
	case MsgReturnToStartOpcode:
		if game == nil {
			player.status = InLobby
			return "", fmt.Errorf(ClientMsgGameGone)
		}
		if !(player.status == InGame && game.gameState == GameOver) {
			return "", fmt.Errorf("player not in game or game not over" + ArgSep + SrvErrInvalidOp)
		}
		player.status = InLobby
		otherPlayer := game.GetOtherPlayer(player)

		if otherPlayer.status == ReadyForGame {
			otherPlayer.status = InLobby
			_, err = sendMsg(otherPlayer.conn, createOpCode(MsgPlayAgainOpcode, false, ClientMsgGameGone), 0)
			if err != nil {
				return "", fmt.Errorf("could not send return to start to player two")
			}
		}
		game.Reset(false)
		removeGame(getGameId(game)) //player left, removing game
		return "left the lobby", nil
	case MsgPingOpcode:
		player.timeSinceLastPing = time.Now()
		return "ping", nil
	case MsgRecoveryOpcode:
		return handleRecoveryOpcode(player, game)
	default:
		return "", fmt.Errorf("unknown opcode")
	}
}

func handleRecoveryOpcode(player *Player, game *TicTacToeGame) (string, error) {
	option := ""
	var err error
	if player.status == InLobby || (game == nil && player.status == InGame) {
		player.status = InLobby //game gone
		option = ClientMsgRecovery_InLobby
	} else if player.status == ReadyForGame {
		option = ClientMsgRecovery_ReadyForGame
	} else if player.status == InGame {
		result := ""
		winner := game.GetGameWinner()
		if winner == nil {
			result = "Draw"
		} else {
			result = winner.name
		}
		otherPlayer := game.GetOtherPlayer(player)
		otherPlayerName := ""
		if otherPlayer.id != 0 {
			otherPlayerName = otherPlayer.name
		}
		board := game.GetBoardInParsableFormat()
		if game.gameState == WaitingForPlayerOneMove && player.id == game.players[0].id {
			option = ClientMsgRecovery_InGame_YourTurn + ArgSep + board + ArgSep + otherPlayerName
		} else if game.gameState == WaitingForPlayerOneMove && player.id == game.players[1].id {
			option = ClientMsgRecovery_InGame_OtherTurn + ArgSep + board + ArgSep + otherPlayerName
		} else if game.gameState == WaitingForPlayerTwoMove && player.id == game.players[1].id {
			option = ClientMsgRecovery_InGame_YourTurn + ArgSep + board + ArgSep + otherPlayerName
		} else if game.gameState == WaitingForPlayerTwoMove && player.id == game.players[0].id {
			option = ClientMsgRecovery_InGame_OtherTurn + ArgSep + board + ArgSep + otherPlayerName
		} else if game.gameState == GameOver {
			option = ClientMsgRecovery_InGame_GameOver + ArgSep + board + ArgSep + result + ArgSep + otherPlayerName
		}
	} else {
		return "", fmt.Errorf("unknown player state")
	}
	if !player.connected {
		player.connected = true
		if game != nil {
			otherPlayer := game.GetOtherPlayer(player)
			if otherPlayer.id != 0 {
				_, err = sendMsg(otherPlayer.conn, createOpCode(MsgContinueOpcode, true, ""), 0)
				if err != nil {
					return "", fmt.Errorf("could not send continue to other player")
				}
			}
		}
		go disconnectHandler(player)
	}
	player.timeSinceLastPing = time.Now()
	return option, nil
}

func informPlayerAboutDisconnect(player *Player) {
	game := findGame(player)
	if game == nil {
		return
	}
	otherPlayer := game.GetOtherPlayer(player)
	if otherPlayer.id == 0 {
		return
	}
	_, err := sendMsg(player.conn, createOpCode(MsgPauseOpcode, true, ""), 0)
	if err != nil {
		fmt.Println("could not send message to client")
		return
	}
}

func connectionCloseHandler(player *Player) {
	for {
		if player.getTimeSinceLastPing() > time.Second*MaxSecondsBeforeDisconnect {
			if player.conn == nil {
				return
			}
			playerConn := (*player.conn)
			if playerConn == nil {
				playerDisconnected(player)
				return
			}
			log.Println("Closing connection due to timeout: ", player.getTimeSinceLastPing())
			err := playerConn.Close()
			if err != nil {
				log.Println("could not close connection")
			}
			return
		}
	}
}

func disconnectHandler(player *Player) {
	for {
		time.Sleep(time.Second * PingTime)
		updatePlayerConnected(player)
		if !player.connected {
			log.Printf("Player %s (ID: %d) disconnected\n", player.name, player.id)
			game := findGame(player)
			if game == nil {
				return
			}
			otherPlayer := game.GetOtherPlayer(player)
			if otherPlayer.id == 0 {
				return
			}
			informPlayerAboutDisconnect(otherPlayer)
			return
		}
	}
}

// check ping duration for player
func updatePlayerConnected(player *Player) {
	if player.conn != nil {
		if player.getTimeSinceLastPing() > time.Second*PingTime*MaxNoPingReceived {
			player.connected = false
		} else {
			player.connected = true
		}
	}
}

// get game id in list of available games
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

// find game that player is in
func findGame(player *Player) *TicTacToeGame {
	gameListMutex.Lock()
	defer gameListMutex.Unlock()
	for _, v := range availableGamesList {
		if v.players[0].id == player.id || v.players[1].id == player.id {
			return v
		}
	}
	//player not in any game
	return nil
}

// join game that is not full
func joinGame(player *Player) *TicTacToeGame {
	gameListMutex.Lock()
	defer gameListMutex.Unlock()
	for i, v := range availableGamesList {
		if v.players[0].id == 0 || v.players[1].id == 0 {
			v.Join(player)
			return availableGamesList[i]
		}
	}
	//no available games
	return nil
}

// create new game
func createGame() *TicTacToeGame {
	newGame := NewTickTackToeGame(defaultBoardSize)
	gameListMutex.Lock()
	defer gameListMutex.Unlock()
	availableGamesList = append(availableGamesList, newGame)
	return newGame
}
