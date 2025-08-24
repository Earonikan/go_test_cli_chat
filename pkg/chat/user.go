package chat

import (
	"time"

	"golang.org/x/net/websocket"
)

type UserClient struct {
	Username   string          `json:"username"`
	Password   string          `json:"password"`
	Connection *websocket.Conn `json:"-"`
	Egress     chan *Message   `json:"-"`
	Supervisor *Supervisor     `json:"-"`
}

func NewUserClient(name string, password string, connection *websocket.Conn, supervisor *Supervisor) *UserClient {
	return &UserClient{
		Username:   name,
		Password:   password,
		Connection: connection,
		Egress:     make(chan *Message),
		Supervisor: supervisor,
	}
}

func (u *UserClient) Read() {
	for {
		message := &Message{}
		if err := websocket.JSON.Receive(u.Connection, message); err != nil {
			// EOF connection closed by the client
			u.Supervisor.Quit(u)
			break
		}

		message.SetTime(time.Now())
		u.Supervisor.Broadcast(message)
	}
}

func (u *UserClient) Write(message *Message) {
	if err := websocket.JSON.Send(u.Connection, message); err != nil {
		// EOF connection closed by the client
		u.Supervisor.Quit(u)
	}
}
