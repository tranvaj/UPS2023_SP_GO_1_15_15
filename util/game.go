package util

import (
	"errors"
	"strconv"
	"strings"
	"sync"
)

type TicTacToeGame struct {
	board          [][]int
	players        [2]*Player
	gameState      int
	gameOverState  int
	readyPlayerOne int
	readyPlayerTwo int
	moveCount      int
	mu             sync.Mutex
}

func (g *TicTacToeGame) Join(player *Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.gameState != WaitingForPlayersReady {
		return errors.New("game already started or over")
	}
	//check if game is full
	if g.players[0].Id == 0 {
		g.players[0] = player
		g.readyPlayerOne = 1
		return nil
	}

	if g.players[1].Id == 0 {
		g.players[1] = player
		g.readyPlayerTwo = 1
		return nil
	}
	return errors.New("max number of players reached")
}

func (g *TicTacToeGame) GetGameWinner() *Player {
	if g.gameOverState == PlayerOneWin {
		return g.players[0]
	} else if g.gameOverState == PlayerTwoWin {
		return g.players[1]
	} else {
		return nil
	}
}

func (g *TicTacToeGame) GetOtherPlayer(player *Player) *Player {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.players[0].Id == player.Id {
		return g.players[1]
	} else {
		return g.players[0]
	}
}

func (g *TicTacToeGame) PlayAgain(player Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.gameState != GameOver {
		return errors.New("game not over")
	}
	if g.players[0].Id == player.Id {
		g.readyPlayerOne = 1
	} else if g.players[1].Id == player.Id {
		g.readyPlayerTwo = 1
	} else {
		return errors.New("player not in game")
	}
	return nil
}

func (g *TicTacToeGame) RemovePlayer(player *Player) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.players[0].Id == player.Id {
		*g.players[0] = Player{}
	} else if g.players[1].Id == player.Id {
		*g.players[1] = Player{}
	}
}

func (g *TicTacToeGame) Move(player Player, x int, y int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	//check if game is over
	if g.gameState == GameOver {
		return errors.New("game is over")
	}
	//check if move valid based on game board length adaptively
	if x < 0 || x >= len(g.board) || y < 0 || y >= len(g.board[0]) {
		return errors.New("invalid move")
	}
	//check if field is already occupied
	if g.board[x][y] != 0 {
		return errors.New("field already occupied")
	}
	//check if player is allowed to move
	if g.gameState == WaitingForPlayerOneMove && player.Id == g.players[0].Id {
		g.board[x][y] = player.Id
	} else if g.gameState == WaitingForPlayerTwoMove && player.Id == g.players[1].Id {
		g.board[x][y] = player.Id
	} else {
		return errors.New("not players turn")
	}
	g.moveCount++

	//change game state
	if g.gameState == WaitingForPlayerOneMove {
		g.gameState = WaitingForPlayerTwoMove
	} else {
		g.gameState = WaitingForPlayerOneMove
	}

	//check win
	win, _ := g.checkWin(player)
	if win {
		if player.Id == g.players[0].Id {
			g.gameOverState = PlayerOneWin
		} else {
			g.gameOverState = PlayerTwoWin
		}
		g.gameState = GameOver
	}
	//check draw
	if g.moveCount == (len(g.board) * len(g.board)) {
		g.gameOverState = Draw
		g.gameState = GameOver
	}
	return nil
}

func (g *TicTacToeGame) GetBoardInParsableFormat() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	var result string
	for _, row := range g.board {
		for i, col := range row {
			if g.players[0].Id == col {
				col = 1
			} else if g.players[1].Id == col {
				col = 2
			} else {
				col = 0
			}
			result += strconv.Itoa(col)
			if i != len(row)-1 {
				result += colSep
			}
		}
		result += rowSep
	}
	result = strings.TrimSuffix(result, rowSep) // Remove the extra semicolon at the end
	return result
}

// check if player won the game on variable board size
func (g *TicTacToeGame) checkWin(player Player) (bool, Player) {
	// check col
	for i := 0; i < len(g.board); i++ {
		win := true
		for j := 0; j < len(g.board); j++ {
			if g.board[i][j] != player.Id {
				win = false
				break
			}
		}
		if win {
			return true, player
		}
	}

	// check row
	for i := 0; i < len(g.board); i++ {
		win := true
		for j := 0; j < len(g.board); j++ {
			if g.board[j][i] != player.Id {
				win = false
				break
			}
		}
		if win {
			return true, player
		}
	}

	// check diagonal
	win := true
	for i := 0; i < len(g.board); i++ {
		if g.board[i][i] != player.Id {
			win = false
			break
		}
	}
	if win {
		return true, player
	}

	// check anti-diagonal
	win = true
	for i := 0; i < len(g.board); i++ {
		if g.board[i][len(g.board)-1-i] != player.Id {
			win = false
			break
		}
	}
	if win {
		return true, player
	}

	return false, Player{}
}

// new tictactoe game
func NewTickTackToeGame(boardSize int) *TicTacToeGame {
	board := make([][]int, boardSize)
	for i := 0; i < boardSize; i++ {
		board[i] = make([]int, boardSize)
	}
	players := [2]*Player{}
	players[0] = &Player{}
	players[1] = &Player{}
	return &TicTacToeGame{
		board:          board,
		players:        players,
		gameState:      WaitingForPlayersReady,
		gameOverState:  NotOver,
		readyPlayerOne: 0,
		readyPlayerTwo: 0,
		moveCount:      0,
	}
}
func (g *TicTacToeGame) IsFull() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.players[0].Id != 0 && g.players[1].Id != 0
}

func (g *TicTacToeGame) IsReady() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.readyPlayerOne != 0 && g.readyPlayerTwo != 0
}

func (g *TicTacToeGame) Start() error {
	if !g.IsFull() {
		return errors.New("game not full")
	}
	if !g.IsReady() {
		return errors.New("players not ready")
	}
	g.Reset(true)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.gameState = WaitingForPlayerOneMove
	return nil
}

func (g *TicTacToeGame) Reset(keepPlayers bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.board = make([][]int, len(g.board))
	for i := 0; i < len(g.board); i++ {
		g.board[i] = make([]int, len(g.board))
	}
	if !keepPlayers {
		g.players = [2]*Player{}
	}
	g.gameState = WaitingForPlayersReady
	g.gameOverState = NotOver
	g.readyPlayerOne = 0
	g.readyPlayerTwo = 0
	g.moveCount = 0
}
