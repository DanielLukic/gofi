package main

import (
	"flag"
	"os"

	"gofi/pkg/gofi2"
	"gofi/pkg/log"
)

func main() {
	logLevel := flag.String("log", "info", "Set logging level (off, error, warning, info, debug)")
	kill := flag.Bool("kill", false, "Kill running gofi2 instance")
	gui := flag.Bool("gui", false, "Use GUI instead of terminal")
	flag.Parse()

	log.SetupLogger(*logLevel, false)

	if *kill {
		log.Debug("Killing gofi2 instance")
		gofi2.KillInstance()
		os.Exit(0)
	}

	instanceManager := gofi2.NewInstanceManager()
	defer instanceManager.Cleanup()

	if instanceManager.CheckExistingInstance() {
		log.Debug("Another instance already running, signaled and exiting")
		os.Exit(0)
	}

	if *gui {
		app := gofi2.NewGUIApp()
		defer app.Cleanup()

		if err := instanceManager.StartIPCServer(app); err != nil {
			log.Error("Failed to start IPC server: %s", err)
			os.Exit(1)
		}

		if err := app.Start(); err != nil {
			log.Error("Failed to start app: %s", err)
			os.Exit(1)
		}

		app.Show()
		app.Run()
	} else {
		app := gofi2.NewApp()
		defer app.Cleanup()

		if err := instanceManager.StartIPCServer(app); err != nil {
			log.Error("Failed to start IPC server: %s", err)
			os.Exit(1)
		}

		if err := app.Start(); err != nil {
			log.Error("Failed to start app: %s", err)
			os.Exit(1)
		}

		app.Show()
		app.Run()
	}
}
