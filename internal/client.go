package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gorilla/websocket"
	"github.com/rivo/tview"
)

var incoming chan Message
var outgoing chan Message

func RunChat(url string) {
	var app = tview.NewApplication()
	nameInput := tview.NewInputField().SetLabel("Name :").SetFieldWidth(20)
	Submit := func() {
		userName := nameInput.GetText()
		if userName == "" {
			return
		}
		ShowChat(app, userName, url)

	}
	form := tview.NewForm().
		AddFormItem(nameInput).
		AddButton("join", Submit).
		AddButton("quit", func() {
			app.Stop()
		})

	form.SetBorder(true).SetTitle("Chat").SetTitleAlign(tview.AlignLeft)

	nameInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			Submit()
		}
	})

	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}

}

func ShowChat(app *tview.Application, username string, url string) {
	incoming = make(chan Message)
	outgoing = make(chan Message)
	chatLog := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	msgInput := tview.NewInputField().SetLabel("Enter Message :").SetFieldWidth(20)
	flex := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(chatLog, 0, 1, false).AddItem(msgInput, 1, 1, true)
	flex.SetTitle("Chat").SetTitleAlign(tview.AlignLeft).SetBorder(true)
	go ClientDial(app, chatLog, url)
	msgInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			data := msgInput.GetText()
			if data == "" {
				return
			}
			Msg := Message{
				Name: username,
				Data: data,
				Time: time.Now().Format("15:04"),
			}
			outgoing <- Msg
			msgInput.SetText("")
		}
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(chatLog)
			return nil
		case tcell.KeyEsc:
			app.SetFocus(msgInput)
		}
		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}

}

func ClientDial(app *tview.Application, chatLog *tview.TextView, url string) {
	Myurl := url
	conn, _, err := websocket.DefaultDialer.Dial(Myurl, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		for {
			var Msg Message
			_, message, err := conn.ReadMessage()
			if err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintln(chatLog, err)
				})
				return
			}
			json.Unmarshal(message, &Msg)
			incoming <- Msg
		}
	}()

	for {
		select {
		case Msg := <-outgoing:
			msgJson, _ := json.Marshal(Msg)
			err = conn.WriteMessage(websocket.TextMessage, msgJson)
			if err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintln(chatLog, err)
				})
			}
		case message := <-incoming:
			Msg := message
			app.QueueUpdateDraw(func() {
				fmt.Fprintln(chatLog, Msg.Time, "[", Msg.Name, "] -->	", Msg.Data)
				chatLog.ScrollToEnd()
			})
		}
	}
}
