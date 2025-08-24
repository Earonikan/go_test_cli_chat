package transport

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lmikolajczak/go-cli-chat/pkg/chat"
	"golang.org/x/net/websocket"

	_ "github.com/lib/pq"
)

type Transport struct {
	interrupt  chan os.Signal
	wg         sync.WaitGroup
	supervisor *chat.Supervisor
	db         *sql.DB
}

func CreateClientApp() *Transport {
	return &Transport{
		make(chan os.Signal),
		sync.WaitGroup{},
		&chat.Supervisor{},
		&sql.DB{},
	}
}

func (t *Transport) OnServerDisconnected() {
	defer t.wg.Done()
	<-t.interrupt

	notification := chat.NewMessage(chat.ServerClosed, "System", "Server has beeen closed...")
	notification.SetTime(time.Now())
	t.supervisor.Broadcast(notification)

	log.Println("Server closed...")
}

func (t *Transport) initDB(supervisor *chat.Supervisor) {
	var err error
	connStr := "host=localhost port=5432 user=postgres password=admin dbname=AuthDB sslmode=disable"
	t.db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = t.db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	supervisor.ConnectToDB(t.db)

	fmt.Println("Connected to the PostgresSQL database!")
}

func (t *Transport) Serve(supervisor *chat.Supervisor, sAddress string) {
	t.interrupt = make(chan os.Signal, 2)
	signal.Notify(t.interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	t.supervisor = supervisor

	t.initDB(supervisor)

	t.wg.Add(1)
	go t.OnServerDisconnected()

	// Use websocket.Server because we want to accept non-browser clients,
	// which do not send an Origin header. websocket.Handler does check
	// the Origin header by default.
	http.Handle("/", websocket.Server{
		Handler: t.supervisor.ServeWS(),
		// Set a Server.Handshake to nil - does not check the origin.
		// We can always provide a custom handshake method to access
		// the handshake http request and implement origin check or
		// other custom logic before the connection is established.
		Handshake: nil,
	})

	go func() {
		log.Printf("Server started on %v...\n", sAddress)
		err := http.ListenAndServe(sAddress, nil)
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	t.wg.Wait()
}
