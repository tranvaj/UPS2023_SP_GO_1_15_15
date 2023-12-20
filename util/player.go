package util

import (
	"errors"
	"net"
	"sync"
	"time"
)

type Player struct {
	Id                int       // 0 means not invalid or empty player
	Name              string    // name of the player
	Conn              *net.Conn // connection of the player
	ClientId          int       // client id of the player (kinda useless but whatever)
	TimeSinceLastPing time.Time // time since last ping
	Status            int       // status of the player
	Connected         bool      // is player connected
}

type Players struct {
	PlayerId int        // id of the player
	Players  []*Player  // slice of players (ptr)
	mu       sync.Mutex // mutex for thread safety
}

// Gets duration since last ping
func (q *Player) getTimeSinceLastPing() time.Duration {
	return time.Since(q.TimeSinceLastPing)
}

func NewPlayers() *Players {
	return &Players{PlayerId: 1, Players: make([]*Player, 0)}
}

func NewPlayer() *Player {
	return &Player{Id: 0, Name: "", Conn: nil, ClientId: 0, TimeSinceLastPing: time.Now(), Status: InLobby, Connected: true}
}

func (q *Players) GetPlayerIndexByName(name string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, v := range q.Players {
		if v.Name == name {
			return i
		}
	}
	return -1
}

func (q *Players) GetPlayerIndexByPtr(player *Player) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, v := range q.Players {
		if v == player {
			return i
		}
	}
	return -1
}

func (q *Players) GetPlayerByClientId(clientId int) *Player {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, v := range q.Players {
		if v.ClientId == clientId {
			return v
		}
	}
	return nil
}

// Login searches for a player with the specified name in the Players and updates their connection information.
// If a player with the specified name is found, their connection, last ping time is updated and the player is returned.
// If no player with the specified name is found, an error is returned.
func (q *Players) Login(conn *net.Conn, name string, player *Player) (*Player, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, v := range q.Players {
		if v.Name == name {
			v.Conn = conn
			v.TimeSinceLastPing = time.Now()
			return v, nil
		}
	}
	return nil, errors.New("player not found")
}

// Does not set player.Conn and client id
// Appends player to the end of the slice
func (q *Players) AddNewPlayer(player *Player) error {
	if q.getPlayersLen() >= MaxClients {
		return errors.New("max number of players reached")
	}
	if q.GetPlayerIndexByName(player.Name) != -1 {
		return errors.New("player with this name already exists")
	}
	if player.Conn == nil {
		return errors.New("player connection is nil")
	}

	player.Id = q.PlayerId
	player.TimeSinceLastPing = time.Now()
	player.Status = InLobby
	player.Connected = true
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Players = append(q.Players, player)
	q.PlayerId++
	return nil
}

// Removes player from Players
func (q *Players) Logout(player *Player) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, v := range q.Players {
		if v.Id == player.Id {
			*q.Players[i] = Player{}
			return
		}
	}
}

func (q *Players) getPlayersLen() int {
	//len(players) is always maxclients because of make
	//so we need to count how many players are actually in the game
	count := 0
	for _, v := range q.Players {
		if v.Id != 0 {
			count++
		}
	}
	return count
}
