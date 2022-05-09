package data

// This module deals with sending messages between the frontend and backend
import (
	"encoding/json"
	"fmt"

	"github.com/Equanox/gotron"
)


var SimActive bool = false

type StringEvent struct {
	*gotron.Event
	Msg     string `json:"Value"`
	VarName string
}

type IntEvent struct {
	*gotron.Event
	Msg     int `json:"Value"`
	VarName string
}


// Message struct for strings
type StringMessage struct {
	EventType  string   `json:"Event"`
	EventValue []string `json:"Value"`
}

// Message struct for UpdatePersonUI
type updateMessage struct {
	*gotron.Event
	EventType  string           `json:"Event"`
	EventValue []UpdatePersonUI `json:"Value"`
}

// Message struct for UILocation
type initLocationMessage struct {
	*gotron.Event
	EventType  string     `json:"Event"`
	EventValue UiLocation `json:"Value"`
}

// Struct used to send information to the UI on persons.
// Only sent for people whose status has changed.
type UpdatePersonUI struct {
	PersonID        int  // which person is affected
	CurrentLocation int  // which location the person should be rendered at
	Infected        bool // if they were infected this time tick, set to true
	Masked          bool // whether person is masked
	Vaccinated      bool // whether person is vaccinated
}

// Struct used to send number of locations and location names
// to the UI, at the beginning of the simulation
type UiLocation struct {
	Num   int
	Names []string
}

type SimulationState struct {
	*gotron.Event
	EventValue bool		`json:"Value"`
}

// Sends a "VarChange" event with a string message to the UI
func SendString(window *gotron.BrowserWindow, msg string, name string) {
	window.Send(&StringEvent{
		Event:   &gotron.Event{Event: "VarChange"},
		Msg:     msg,
		VarName: name,
	})
}

// Sends a "VarChange" event with an int message to the UI
func SendInt(window *gotron.BrowserWindow, msg int, name string) {
	window.Send(&IntEvent{
		Event:   &gotron.Event{Event: "VarChange"},
		Msg:     msg,
		VarName: name,
	})
}

// Sends a "uiUpdate" event with a []UpdatePersonUI message to the UI
func SendUIUpdate(window *gotron.BrowserWindow, msg []UpdatePersonUI) {
	window.Send(&updateMessage{
		Event:      &gotron.Event{Event: "uiUpdate"},
		EventType:  "uiUpdate",
		EventValue: msg,
	})
}

// Sends a "initLocations" event with a UiLocation message to the UI
func SendUILocations(window *gotron.BrowserWindow, msg UiLocation) {
	window.Send(&initLocationMessage{
		Event:      &gotron.Event{Event: "initLocations"},
		EventType:  "initLocations",
		EventValue: msg,
	})
}

// 
func SendSimulationDone(window *gotron.BrowserWindow, msg string, name string) {
	window.Send(&StringEvent{
		Event:   &gotron.Event{Event: "simulationDone"},
		Msg:     msg,
		VarName: name,
	})
}

func SendSimulationState(window *gotron.BrowserWindow, state bool) {
	window.Send(&SimulationState{
		Event:   		&gotron.Event{Event: "simulationState"},
		EventValue:		state,
	})
}

func Receive(window *gotron.BrowserWindow, eventName string, msg *StringMessage, done chan bool) {
	window.On(&gotron.Event{Event: eventName}, func(bin []byte) {
		err := json.Unmarshal(bin, msg)

		if err != nil {
			panic(err)
		}
		done <- true
	})
}

func ReceiveAndBlock(window *gotron.BrowserWindow, eventName string, msg *StringMessage) {
	ch := make(chan bool)
	go Receive(window, eventName, msg, ch)
	<-ch
}

func ReceiveAndBlockLoop(window *gotron.BrowserWindow, eventName string, msg *StringMessage) {
	for {
		ch := make(chan bool)
		go Receive(window, eventName, msg, ch)
		<-ch
	}
}

func MessageReciever(window *gotron.BrowserWindow, eventName string, f func([]string)) {
	for {
		m := StringMessage{}
		ReceiveAndBlock(window, eventName, &m)
		f(m.EventValue)
	}
}

func WaitForButtonPress(window *gotron.BrowserWindow, buttonName string, f func()) {
	for {
		m := StringMessage{}
		ReceiveAndBlock(window, "btnPress", &m)
		fmt.Println(m.EventValue[0])
		if len(m.EventValue) > 0 {
			if m.EventValue[0] == buttonName {
				f()
			}
		}
	}
}

func WaitForButtonPressApply(window *gotron.BrowserWindow, resetChEventloop chan bool, resetChRun chan bool, f func()) {
	for {
		m := StringMessage{}
		ReceiveAndBlock(window, "Apply", &m)
		func() {
			if SimActive {
			fmt.Println(SimActive)
			resetChEventloop <- true
			resetChRun <- true
			}
		}()
		go f()
	}
}

func WaitForButtonPressRun(window *gotron.BrowserWindow, resetChRun chan bool, f func()) {
	for {
		m := StringMessage{}
		ReceiveAndBlock(window, "Run", &m)
		SendSimulationState(window, true)
		f()
		<-resetChRun
	}
}
