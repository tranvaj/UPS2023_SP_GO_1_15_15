package util

import (
	"errors"
	"net"
	"sync"
	"time"
)

type Player struct {
	id                int
	name              string
	conn              *net.Conn
	timeSinceLastPing time.Time
	status            int
	connected         bool
}

type Players struct {
	playerId int
	players  []*Player
	mu       sync.Mutex
}

func (q *Player) getTimeSinceLastPing() time.Duration {
	return time.Since(q.timeSinceLastPing)
}

func NewPlayers() *Players {
	return &Players{playerId: 1, players: make([]*Player, 0)}
}

func (q *Players) GetPlayerIndex(name string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, v := range q.players {
		if v.name == name {
			return i
		}
	}
	return -1
}

func (q *Players) Login(conn *net.Conn, name string) (*Player, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, v := range q.players {
		if v.name == name {
			return v, nil
		}
	}

	if q.getPlayersLen() >= MaxClients {
		return nil, errors.New("max number of players reached")
	}

	newPlayer := Player{id: q.playerId, name: name, conn: conn, timeSinceLastPing: time.Now(), status: InLobby, connected: true}
	q.players = append(q.players, &newPlayer)
	q.playerId++
	return &newPlayer, nil
}

func (q *Players) Logout(player *Player) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, v := range q.players {
		if v.id == player.id {
			*q.players[i] = Player{}
			return
		}
	}
}

func (q *Players) getPlayersLen() int {
	//len(players) is always maxclients because of make
	//so we need to count how many players are actually in the game
	count := 0
	for _, v := range q.players {
		if v.id != 0 {
			count++
		}
	}
	return count
}
