package main

import (
	"code-runner/config"
	"code-runner/server"
	"code-runner/services/codeRunner"
	"code-runner/services/container"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx := context.Background()
	configManager := config.NewConfigManager("./config.json")
	configManager.ReadConfig()

	containerService := container.NewService()
	cr := codeRunner.NewService(containerService)
	s, err := server.NewServer(8080, "localhost")
	if err != nil {
		log.Fatalf("could not start init server: %s\n", err)
	}

	s.CodeRunner = cr
	go s.Run()

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt)

	<-done
	go func() {
		spinner := []string{"|", "/", "–", "\\", "|", "/", "–", "\\"}
		counter := 0
		for {
			if counter%len(spinner) == 0 {
				counter = 0
			}
			fmt.Printf("%s shutting down code-runner...", spinner[counter])
			time.Sleep(125 * time.Millisecond)
			fmt.Print("\n\033[1A\033[K")
			counter++
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	containerService.Shutdown(ctx)
}
