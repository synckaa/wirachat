package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
)

var incoming = make(chan Message)
var outgoing = make(chan Message)
var form *tview.Form

func RunChat(url string) {
	var app = tview.NewApplication()
	nameInput := tview.NewInputField().SetLabel("Name :").SetFieldWidth(30).SetFieldBackgroundColor(tcell.ColorWhite)
	form = tview.NewForm().
		AddFormItem(nameInput).
		AddButton("join", func() {
			userName := nameInput.GetText()
			if userName == "" {
				modal := tview.NewModal().
					SetText("Please input your name").
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						app.SetRoot(form, true)
					})
				app.SetRoot(modal, true)
				return
			}
			ShowChat(app, userName, url)
		}).
		AddButton("quit", func() {
			app.Stop()
		})

	form.SetBorder(true).SetTitle("WIRAChat").SetTitleAlign(tview.AlignCenter)
	form.SetFieldBackgroundColor(tcell.ColorBlack).SetButtonBackgroundColor(tcell.ColorBlack).SetButtonTextColor(tcell.ColorWhite)
	grid := tview.NewGrid(). // atas, tengah, bawah (0 = fleksibel)
					SetColumns(0, 0, 0).                  // kiri, tengah, kanan
					AddItem(form, 0, 1, 1, 1, 0, 0, true) // taruh form di tengah

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}

}

func ShowChat(app *tview.Application, username string, url string) {
	chatLog := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	clientInput := tview.NewInputField().SetLabel("Enter Message :").SetFieldWidth(20).SetFieldBackgroundColor(tcell.ColorBlack)
	flex := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(chatLog, 0, 1, false).AddItem(clientInput, 1, 1, true)
	flex.SetTitle("Chat").SetTitleAlign(tview.AlignLeft).SetBorder(true)
	go ClientDial(app, chatLog, url)
	clientInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			data := clientInput.GetText()
			if data == "" {
				return
			}
			clientMsg := Message{
				Name: username,
				Data: data,
				Time: time.Now().Format("15:04"),
			}
			outgoing <- clientMsg
			clientInput.SetText("")
		}
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(chatLog)
			return nil
		case tcell.KeyEsc:
			app.SetFocus(clientInput)
		}
		return event
	})

	app.SetRoot(flex, true)

}

func ClientDial(app *tview.Application, chatLog *tview.TextView, url string) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		for {
			var serverMsg Message
			_, message, err := conn.ReadMessage()
			if err != nil {
				app.QueueUpdateDraw(func() {
					_, err = fmt.Fprintln(chatLog, err)
					if err != nil {
						panic(err)
					}
				})
				return
			}
			err = json.Unmarshal(message, &serverMsg)
			if err != nil {
				panic(err)
			}
			incoming <- serverMsg
		}
	}()

	for {
		select {
		case clientMsg := <-outgoing:
			clientMsgJson, _ := json.Marshal(clientMsg)
			err = conn.WriteMessage(websocket.TextMessage, clientMsgJson)
			if err != nil {
				app.QueueUpdateDraw(func() {
					_, err = fmt.Fprintln(chatLog, err)
					if err != nil {
						panic(err)
					}
				})
			}
		case serverMsg := <-incoming:
			app.QueueUpdateDraw(func() {
				_, err = fmt.Fprintln(chatLog, serverMsg.Time, "[", serverMsg.Name, "] -->	", serverMsg.Data)
				if err != nil {
					panic(err)
				}
				chatLog.ScrollToEnd()
			})
		}
	}
}
