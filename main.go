package main

import (
	"code-runner/config"
	"code-runner/server"
	"code-runner/services/codeRunner"
	"code-runner/services/container"
	"code-runner/services/scheduler"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx := context.Background()
	var configManager *config.ConfigManager
	if configPath := os.Getenv("CODE_RUNNER_CONFIG"); configPath != "" {
		configManager = config.NewConfigManager(configPath)
	} else {
		configManager = config.NewConfigManager("/etc/code-runner/config.json")
	}
	configManager.ReadConfig()

	containerService := container.NewService()
	schedulerService := scheduler.NewScheduler(time.Second)

	cr := codeRunner.NewService(ctx, containerService, schedulerService)

	s, err := server.NewServer(8080, "0.0.0.0")
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
	s.CodeRunner.Shutdown(ctx)
}
