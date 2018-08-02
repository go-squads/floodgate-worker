package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-squads/floodgate-worker/analytic"
	"github.com/urfave/cli"
)

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

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(fmt.Sprintf("Some error occurred: %s", err.Error()))
	}
}

func ActionAnalyserService(cli *cli.Context) (err error) {
	brokers := []string{
		"localhost:9092",
	}

	analyticService := analytic.NewAnalyticServices(brokers)
	analyticService.Start()
	// timekit.Sleep("100s")
	return
}
