package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/go-squads/floodgate-worker/analytic"
	"github.com/urfave/cli"
)

var stopSig = make(chan os.Signal)

func main() {

	app := cli.App{
		Name:    "analyser-services",
		Usage:   "Provide kafka producer or consumer for Barito project",
		Version: "0.1.0",
		Commands: []cli.Command{
			{
				Name:      "analyser",
				ShortName: "aw",
				Usage:     "start barito-flow stream analyser",
				Action:    ActionAnalyserService,
			},
		},
	}

	signal.Notify(stopSig, os.Interrupt)

	go func() {
		select {
		case sig := <-stopSig:
			fmt.Printf("Got %s signal. Aborting...\n", sig)
			os.Exit(1)
		}
	}()

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(fmt.Sprintf("Some error occurred: %s", err.Error()))
	}
}

func ActionAnalyserService(cli *cli.Context) {
	brokers := []string{
		"localhost:9092",
	}

	analyticService := analytic.NewAnalyticServices(brokers)
	analyticService.Start()
	for {
		select {
		case <-stopSig:
			analyticService.Close()
			return
		}
	}
}
