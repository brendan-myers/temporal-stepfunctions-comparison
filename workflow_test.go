package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/testsuite"
)

var ts, _ = time.Parse("2006-01-02", "2025-01-01")

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}

	scenarios := map[string]int{
		"buy":  100,
		"sell": 0,
	}

	for tradeType, stockPrice := range scenarios {
		expectedResult := TransactionResult{
			Id:        "abcd1234",
			Price:     stockPrice,
			Type:      tradeType,
			Quantity:  10,
			Timestamp: ts,
		}

		env := testSuite.NewTestWorkflowEnvironment()

		env.OnActivity(CheckStockPrice, mock.Anything).Return(stockPrice, nil)
		env.OnActivity(SendNotification, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		env.OnActivity(PerformTrade, mock.Anything, stockPrice, tradeType).Return(expectedResult, nil)

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow("approve", true)
		}, time.Hour)
		env.ExecuteWorkflow(StockTrade)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		var result TransactionResult
		require.NoError(t, env.GetWorkflowResult(&result))

		expectedResult.Approved = true
		require.Equal(t, expectedResult, result)
	}
}

func Test_PerformTrade(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	env.RegisterActivity(PerformTrade)

	stockPrice := 75
	tradeType := "buy"

	val, err := env.ExecuteActivity(PerformTrade, stockPrice, tradeType)
	require.NoError(t, err)

	var result TransactionResult
	require.NoError(t, val.Get(&result))
	result.Timestamp = ts // lazy testing of non-deterministic value
	result.Quantity = 10  // lazy testing of non-deterministic value
	require.Equal(t, result, TransactionResult{
		Id:        "abcd1234",
		Price:     stockPrice,
		Type:      tradeType,
		Quantity:  10,
		Timestamp: ts,
	})
}
