package main

import (
	"github.com/Equanox/gotron"
	"github.com/ospp-projects-2021/clockwork/src/backend"
	"github.com/ospp-projects-2021/clockwork/src/data"
)

func main() {
	// Create a new browser window instance
	window, err := gotron.New("./frontend/webapp")
	if err != nil {
		panic(err)
	}

	// Alter default window size and window title.
	window.WindowOptions.Width = 1680
	window.WindowOptions.Height = 1050
	window.WindowOptions.Title = "Go Virus, Go!"

	// Start the browser window.
	// This will establish a golang <=> nodejs bridge using websockets,
	// to control ElectronBrowserWindow with our window object.
	done, err := window.Start()
	if err != nil {
		panic(err)
	}

	window.OpenDevTools()

	m := data.StringMessage{}
	resetChEventloop := make(chan bool)
	resetChRun := make(chan bool)

	data.ReceiveAndBlock(window, "Ready", &m)
	go data.ReceiveAndBlockLoop(window, "sendSettings", &m)
	go data.WaitForButtonPressApply(window, resetChEventloop, resetChRun, func() {
		settings(window, m)
	})

	go data.WaitForButtonPressRun(window, resetChRun, func() {
		go backend.EventLoop(window, resetChEventloop)
	})
	<-done
}

func settings(window *gotron.BrowserWindow, m data.StringMessage) {
	backend.ReadSettings(&m)
	backend.InitiateSimulation(window)
}
