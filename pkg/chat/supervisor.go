package chat

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"slices"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/websocket"
)

type Supervisor struct {
	UserClients []*UserClient
	mu          sync.Mutex
	db          *sql.DB
}

type ret struct {
	success bool
	msg     string
}

func NewSupervisor() *Supervisor {
	return &Supervisor{
		UserClients: make([]*UserClient, 0),
		mu:          sync.Mutex{},
		db:          &sql.DB{},
	}
}

func (s *Supervisor) ConnectToDB(db *sql.DB) {
	s.db = db
}

func (s *Supervisor) Join(userClient *UserClient) {
	s.mu.Lock()

	s.UserClients = append(s.UserClients, userClient)

	s.mu.Unlock()

	notification := NewMessage(UserClientList, "System", s.CurrentUserClients())
	notification.SetTime(time.Now())
	s.Broadcast(notification)

	notification = NewMessage(Connected, "System", s.LastConnectedUserClient())
	notification.SetTime(time.Now())
	s.Broadcast(notification)

	log.Printf("%v connected...\n", userClient.Username)
}

func (s *Supervisor) Quit(userClient *UserClient) {
	s.mu.Lock()

	for i := len(s.UserClients) - 1; i >= 0; i-- {
		if s.UserClients[i] == userClient {
			s.UserClients = slices.Delete(s.UserClients, i, i+1)
		}
	}

	s.mu.Unlock()

	notification := NewMessage(UserClientList, "System", s.CurrentUserClients())
	notification.SetTime(time.Now())
	s.Broadcast(notification)

	notification = NewMessage(Disconnected, "System", userClient.Username+" disconnected\n")
	notification.SetTime(time.Now())
	s.Broadcast(notification)

	log.Printf("%v has left chat...\n", userClient.Username)
}

func (s *Supervisor) CurrentUserClients() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var userClients string
	for _, userClient := range s.UserClients {
		userClients += fmt.Sprintf("%s\n", userClient.Username)
	}

	return userClients
}

func (s *Supervisor) LastConnectedUserClient() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.UserClients[len(s.UserClients)-1].Username + " connected\n"
}

func (s *Supervisor) Broadcast(message *Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, userClient := range s.UserClients {
		userClient.Write(message)
	}

	return nil
}

func (s *Supervisor) ServeWS() func(connection *websocket.Conn) {
	return func(connection *websocket.Conn) {
		logType := connection.Request().Header.Get("logType")
		userName := connection.Request().Header.Get("Username")
		passWord := connection.Request().Header.Get("Password")
		email := connection.Request().Header.Get("Email")

		var r ret
		userClient := NewUserClient(
			userName,
			passWord,
			connection, s)

		if logType == "-L" {
			userName := connection.Request().Header.Get("Username")
			r = s.Login(userName, passWord)
		} else {
			userName := connection.Request().Header.Get("Username")
			r = s.SignUp(userName, passWord, email)
		}

		if r.success {
			s.Join(userClient)
			userClient.Read()
		} else {
			message := NewMessage(Disconnected, "System", r.msg)
			userClient.Write(message)
		}
	}
}

func (s *Supervisor) Login(username, password string) ret {

	// Retrieve user from the database
	var storedHash string
	query := `SELECT password_hash FROM users WHERE username = $1`
	err := s.db.QueryRow(query, username).Scan(&storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println(err)
			return ret{false, "User not found"}
		}
		log.Println(err)
		return ret{false, "Login query failed"}
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		log.Println(err)
		return ret{false, err.Error()}
	}
	return ret{true, "Login success"}
}

func (s *Supervisor) SignUp(username, password, email string) ret {

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return ret{false, "Problem with password hashing"}
	}

	// Insert user into the database
	query := `INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)`
	_, err = s.db.Exec(query, username, string(hashedPassword), email)
	if err != nil {
		log.Println(err)
		return ret{false, "User already exist, try to login"}
	}
	return ret{true, "Sign-up success"}
}
