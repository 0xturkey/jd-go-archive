package utils

import (
	"errors"
	"fmt"
	"log"

	"github.com/0xturkey/jd-go/model"
)

type Task model.Task

type TaskHandler interface {
	RunTask(encodedTx string, rpcURL string) (string, error)
	GetTaskStatus(taskID string, rpcURL string) (model.TaskStatus, error)
}

type EthTx struct{}

func (e *EthTx) RunTask(encodedTx string, rpcURL string) (string, error) {
	if ok, err := preTransactionCheck(encodedTx, rpcURL); !ok {
		return "", err
	}

	respObj, err := JSONRPCRequest(rpcURL, "eth_sendRawTransaction", []string{encodedTx}, 1)
	if err != nil {
		return "", err
	}

	result, ok := respObj.Result.(string)
	if !ok {
		return "", errors.New("result is not a string")
	}

	return result, nil
}

func (e *EthTx) GetTaskStatus(taskID string, rpcURL string) (model.TaskStatus, error) {
	resp, err := JSONRPCRequest(rpcURL, "eth_getTransactionReceipt", []string{taskID}, 1)
	if err != nil {
		return "", err
	}

	// Check if the response result is nil, which means the transaction is not yet mined.
	if resp.Result == nil {
		return model.InProgress, nil
	}

	log.Printf("Transaction receipt: %+v\n", resp.Result)

	receipt, ok := resp.Result.(map[string]interface{})
	if !ok {
		return model.Fail, errors.New("getTransactionReceipt result is not a map")
	}

	log.Printf("Transaction receipt: %+v\n", receipt)

	// The status is a hex code: 0x1 for success and 0x0 for failure.
	status, isOk := receipt["status"].(string)
	if !isOk {
		return model.Fail, errors.New("status field is missing or not a string")
	}

	if status == "0x1" {
		return model.Completed, nil
	} else if status == "0x0" {
		return model.Fail, nil
	} else {
		return model.Fail, fmt.Errorf("unknown status code: %s", status)
	}
}

// Mock function to simulate getting a task handler.
// You'll need to implement the actual logic.
func GetTaskHandler(taskType model.TaskType) (TaskHandler, error) {
	switch taskType {
	case model.EthTx:
		return &EthTx{}, nil
	default:
		return nil, errors.New("unsupported task type")
	}
}
