package service

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/sevlyar/go-daemon"

	"risp/config"
	"risp/engine"
)

func handleSigTerm(sig os.Signal) (err error) {
	log.Printf("Should stop now\n")

	// stop <- struct{}{}
	// if sig == syscall.SIGQUIT {
	// 	<-done
	// }

	return daemon.ErrStop
}

type Service struct {
	Config        *config.Config
	DaemonContext *daemon.Context
	DaemonProcess *os.Process
	Engine        *engine.Engine
	NoExit        bool
	Signal        string
}

func NewService(config *config.Config) *Service {
	return &Service{
		Config: config,
	}
}

func (service *Service) Run() (err error) {
	service.Engine = engine.NewEngine(service.Config)

	if service.NoExit {
		if err = service.Engine.Start(); err != nil {
			return
		}

		return
	}

	daemon.AddCommand(daemon.StringFlag(&service.Signal, "stop"), syscall.SIGTERM, handleSigTerm)

	var daemonProcess *os.Process

	service.DaemonContext = &daemon.Context{
		WorkDir:     "./",
		PidFileName: service.Config.PIDFilePath,
		LogFileName: service.Config.LogFilePath,
		PidFilePerm: 0644,
		LogFilePerm: 0640,
		Umask:       027,
		Args:        []string{},
	}

	if daemonProcess, err = service.DaemonContext.Reborn(); err != nil {
		log.Fatal(err)
	}

	if daemonProcess != nil {
		// parent process
		return
	}

	// daemon process

	return service.start()
}

func (service *Service) start() (err error) {
	defer func() {
		if err := service.stop(); err != nil {
			log.Fatal(err)
		}
	}()

	if err = service.Engine.Start(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("About to serve signals\n")

	err = daemon.ServeSignals()
	fmt.Printf("Signals no more, err=%+v\n", err)
	return
}

func (service *Service) stop() (err error) {
	if err := service.DaemonContext.Release(); err != nil {
		return fmt.Errorf("failed to release .pid file: %+v", err)
	}

	return
}

// func (service *Service) restart() (err error) {
// 	return
// }
