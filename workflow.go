package common

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

type TransactionResult struct {
	Approved  bool
	Id        string
	Price     int
	Type      string
	Quantity  int
	Timestamp time.Time
}

func StockTrade(ctx workflow.Context) (TransactionResult, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Second,
	})

	log := workflow.GetLogger(ctx)

	//
	//	Get stock price
	//
	var stockPrice int
	err := workflow.ExecuteActivity(ctx, CheckStockPrice).Get(ctx, &stockPrice)
	if err != nil {
		log.Error("ExecuteActivity failed: %v", err)
		return TransactionResult{}, err
	}

	//
	//	Make recommendation
	//
	var recommendation string
	if stockPrice > 50 {
		recommendation = "buy"
	} else {
		recommendation = "sell"
	}

	//
	//	Send notification
	//
	msg := fmt.Sprintf("Trade %s requires approval", workflow.GetInfo(ctx).WorkflowExecution.ID)
	workflow.ExecuteActivity(ctx, SendNotification, "humanApproval", msg)

	//
	//	Wait for human approval
	//
	var approved bool
	workflow.GetSignalChannel(ctx, "approve").Receive(ctx, &approved)

	if !approved {
		msg = fmt.Sprintf("Trade %s not approved", workflow.GetInfo(ctx).WorkflowExecution.ID)
		workflow.ExecuteActivity(ctx, SendNotification, "report", msg)
		return TransactionResult{Approved: false}, nil
	}

	//
	//	Buy or sell
	//
	var transactionResult TransactionResult

	err = workflow.ExecuteActivity(ctx, PerformTrade, stockPrice, recommendation).Get(ctx, &transactionResult)
	if err != nil {
		log.Error("ExecuteActivity failed: %v", err)
		return TransactionResult{}, err
	}
	transactionResult.Approved = true

	//
	//	Report result
	//
	msg = fmt.Sprintf("Trade %s: %v", workflow.GetInfo(ctx).WorkflowExecution.ID, transactionResult)
	workflow.ExecuteActivity(ctx, SendNotification, "report", msg).Get(ctx, nil)

	return transactionResult, nil
}

func CheckStockPrice(ctx context.Context) (int, error) {
	stockPrice := rand.Intn(100)
	return stockPrice, nil
}

func SendNotification(ctx context.Context, channel string, message string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Simulating notification", channel, message)
	return nil
}

func PerformTrade(ctx context.Context, stockPrice int, tradeType string) (TransactionResult, error) {
	transactionResult := TransactionResult{
		Id:        "abcd1234",
		Price:     stockPrice,
		Type:      tradeType,
		Quantity:  rand.Intn(10),
		Timestamp: time.Now(),
	}

	return transactionResult, nil
}
