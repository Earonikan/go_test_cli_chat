package frontend

import (
	"fmt"
	"sync"

	"github.com/jroimartin/gocui"
	"golang.org/x/net/websocket"

	"github.com/lmikolajczak/go-cli-chat/pkg/chat"
)

const (
	WebsocketEndpoint = "ws://10.227.155.86:8080/"
	WebsocketOrigin   = "http://"

	MessageWidget = "messages"
	UsersWidget   = "users"
	InputWidget   = "send"
)

type chatUI struct {
	*gocui.Gui

	username   string
	password   string
	Email      string
	connection *websocket.Conn
	interupt   chan bool
	wg         sync.WaitGroup
}

func NewChatUI() (*chatUI, error) {
	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, fmt.Errorf("NewChatUI: %w", err)
	}

	return &chatUI{Gui: gui, interupt: make(chan bool), wg: sync.WaitGroup{}}, nil
}

func (ui *chatUI) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	g.Cursor = true

	if messages, err := g.SetView(MessageWidget, 0, 0, maxX-20, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		messages.Title = MessageWidget
		messages.Autoscroll = true
		messages.Wrap = true
	}

	if input, err := g.SetView(InputWidget, 0, maxY-5, maxX-20, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		input.Title = InputWidget
		input.Autoscroll = false
		input.Wrap = true
		input.Editable = true
	}

	if users, err := g.SetView(UsersWidget, maxX-20, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		users.Title = UsersWidget
		users.Autoscroll = false
		users.Wrap = true
	}

	g.SetCurrentView(InputWidget)

	return nil
}

func (ui *chatUI) SetKeyBindings(g *gocui.Gui) error {
	if err := g.SetKeybinding(InputWidget, gocui.KeyCtrlC, gocui.ModNone, ui.Quit); err != nil {
		return err
	}

	if err := g.SetKeybinding(InputWidget, gocui.KeyEnter, gocui.ModNone, ui.WriteMessage); err != nil {
		return err
	}

	if err := g.SetKeybinding(MessageWidget, gocui.KeyArrowUp, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			ui.scrollView(v, -1)
			return nil
		}); err != nil {
		return err
	}

	if err := g.SetKeybinding(MessageWidget, gocui.KeyArrowDown, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			ui.scrollView(v, 1)
			return nil
		}); err != nil {
		return err
	}

	return nil
}

func (ui *chatUI) scrollView(v *gocui.View, dy int) error {
	if v != nil {
		v.Autoscroll = false
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+dy); err != nil {
			return err
		}
	}
	return nil
}

func (ui *chatUI) SetUsername(username string) {
	ui.username = username
}

func (ui *chatUI) SetPassword(password string) {
	ui.password = password
}

func (ui *chatUI) SetEmail(email string) {
	ui.Email = email
}

func (ui *chatUI) SetConnection(connection *websocket.Conn) {
	ui.connection = connection
}

func (ui *chatUI) Connect(logType string) error {
	config, err := websocket.NewConfig(WebsocketEndpoint, WebsocketOrigin)
	if err != nil {
		return err
	}

	config.Header.Set("Username", ui.username)
	config.Header.Set("Password", ui.password)
	config.Header.Set("Email", ui.Email)
	config.Header.Set("logType", logType)

	connection, err := websocket.DialConfig(config)
	if err != nil {
		return err
	}

	ui.SetConnection(connection)

	return nil
}

func (ui *chatUI) WriteMessage(_ *gocui.Gui, v *gocui.View) error {
	message := chat.NewMessage(chat.Regular, ui.username, v.Buffer())

	if err := websocket.JSON.Send(ui.connection, message); err != nil {
		return fmt.Errorf("chatUI.WriteMessage: %w", err)
	}

	v.SetCursor(0, 0)
	v.Clear()

	return nil
}

func (ui *chatUI) ReadMessage() error {
	for {
		var message chat.Message
		if err := websocket.JSON.Receive(ui.connection, &message); err != nil {
			return fmt.Errorf("chatUI.ReadMessage: %w", err)
		}

		ui.Update(func(g *gocui.Gui) error {
			switch message.Type {
			case chat.Regular, chat.Connected, chat.Disconnected, chat.ServerClosed:
				view, err := ui.View(MessageWidget)
				if err != nil {
					return fmt.Errorf("chatUI.ReadMessage: %w", err)
				}

				fmt.Fprint(view, message.Formatted())

			case chat.UserClientList:
				view, err := ui.View(UsersWidget)
				if err != nil {
					return fmt.Errorf("chatUI.ReadMessage: %w", err)
				}

				view.Clear()
				fmt.Fprint(view, message.Text)

				// case chat.Connected, chat.Disconnected, chat.ServerClosed:

				// 	view, err := ui.View(MessageWidget)
				// 	if err != nil {
				// 		return fmt.Errorf("chatUI.ReadMessage: %w", err)
				// 	}
				// 	fmt.Fprint(view, message.Formatted())

			}

			return nil
		})
	}
}

func (ui *chatUI) Quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

// func (ui *chatUI) OnServerDisconnected() {
// 	defer ui.wg.Done()
// 	<-ui.interupt
// 	log.Println("Client has been closed...")
// }

func (ui *chatUI) Serve() error {
	go ui.ReadMessage()

	// ui.wg.Add(1)
	// go ui.OnServerDisconnected()

	if err := ui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return fmt.Errorf("chatUI.Serve: %w", err)
	}

	return nil
}
