package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

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

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(fmt.Sprintf("Some error occurred: %s", err.Error()))
	}
}

func ActionAnalyserService(cli *cli.Context) {
	analytic.LoadEnviromentConfig()
	var brokers []string
	for _, address := range strings.Split(os.Getenv("BROKERS_ADDRESS"), ",") {
		fmt.Println(address)
		brokers = append(brokers, address)
	}

	analyticService := analytic.NewAnalyticServices(brokers)
	err := analyticService.Start()
	if err != nil {
		log.Fatal("Failed to start")
	}

	for {
		select {
		case <-stopSig:
			fmt.Println("Gracefully shutting down..")
			analyticService.Close()
			os.Exit(1)
		}
	}
}
