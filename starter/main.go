package main

import (
	"context"
	"log"
	"os"
	common "temporal/stepfunctions-orchestration"

	"go.temporal.io/sdk/client"
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

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: common.TASK_QUEUE,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, common.StockTrade)
	if err != nil {
		log.Fatalln("Failed to execute workflow", err)
	}

	log.Println("Workflow started", "id", we.GetID(), "runId", we.GetRunID())

	var result common.TransactionResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Failed to get workflow result", err)
	}

	log.Println("Workflow result: ", result)
}
