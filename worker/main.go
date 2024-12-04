package main

import (
	"log"
	"os"
	common "temporal/stepfunctions-orchestration"

	"go.temporal.io/sdk/worker"
)

func main() {
	params, err := common.ParseParams(os.Args[1:])
	if err != nil {
		log.Fatalln("Failed to parse input parameters", err)
	}

	c, err := common.Connect(&params)
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	w := worker.New(c, common.TASK_QUEUE, worker.Options{})

	w.RegisterWorkflow(common.StockTrade)
	w.RegisterActivity(common.CheckStockPrice)
	w.RegisterActivity(common.SendNotification)
	w.RegisterActivity(common.PerformTrade)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start Worker", err)
	}
}
