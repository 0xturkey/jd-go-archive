package utils

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

func preTransactionCheck(encodedTx string, rpcURL string) (bool, error) {
	// Remove the "0x" prefix
	encodedTx = strings.TrimPrefix(encodedTx, "0x")

	rawTx, err := hex.DecodeString(encodedTx)
	if err != nil {
		return false, err
	}

	log.Printf("Decoded transaction: %v\n", rawTx)

	// Decode the RLP-encoded transaction
	tx := &types.Transaction{}
	if err := tx.UnmarshalBinary(rawTx); err != nil {
		return false, fmt.Errorf("rlp: failed to decode transaction: %v", err)
	}

	log.Printf("Transaction: %+v\n", tx)

	sender, err := checkSignatureAndGetSigner(tx)
	if err != nil {
		return false, err
	}

	log.Printf("Sender: %s\n", sender.Hex())

	type checkResult struct {
		Valid bool
		Error error
	}

	// Create channels for the results of the checks
	nonceChan := make(chan checkResult, 1)
	gasChan := make(chan checkResult, 1)

	// Check nonce in a goroutine
	go func() {
		ok, err := isNonceValid(rpcURL, sender, tx.Nonce())
		nonceChan <- checkResult{Valid: ok, Error: err}
	}()

	// Check gas estimation in a goroutine
	go func() {
		ok, err := checkGasEstimation(rpcURL, tx, sender)
		gasChan <- checkResult{Valid: ok, Error: err}
	}()

	// Wait for the nonce check to complete
	nonceResult := <-nonceChan
	if nonceResult.Error != nil || !nonceResult.Valid {
		return false, nonceResult.Error
	}

	// Wait for the gas estimation check to complete
	gasResult := <-gasChan
	if gasResult.Error != nil || !gasResult.Valid {
		return false, gasResult.Error
	}

	return true, nil
}

func checkSignatureAndGetSigner(tx *types.Transaction) (common.Address, error) {
	log.Printf("Checking signature for transaction: %+v\n", tx.Protected())

	signer := types.LatestSignerForChainID(tx.ChainId())
	sender, err := types.Sender(signer, tx)
	if err != nil {
		return common.Address{}, err
	}

	return sender, nil
}

func isNonceValid(rpcURL string, sender common.Address, nonce uint64) (bool, error) {
	// Get the current nonce for the sender
	respObj, err := JSONRPCRequest(rpcURL, "eth_getTransactionCount", []interface{}{sender.Hex(), "latest"}, 1)
	if err != nil {
		return false, err
	}

	// Parse the result to uint64
	var currentNonceStr string
	var ok bool
	if currentNonceStr, ok = respObj.Result.(string); !ok {
		return false, errors.New("nonce is not a string")
	}
	currentNonce, err := hexutil.DecodeUint64(currentNonceStr)
	if err != nil {
		return false, err
	}

	if nonce != currentNonce {
		return false, fmt.Errorf("nonce mismatch: expected %d, got %d", nonce, currentNonce)
	}

	return true, nil
}

func checkGasEstimation(rpcURL string, tx *types.Transaction, sender common.Address) (bool, error) {
	reqData := map[string]interface{}{
		"from":  sender.Hex(),
		"to":    tx.To(),
		"value": "0x" + tx.Value().Text(16),
		"data":  tx.Data(),
	}

	// Get the gas estimation for the transaction
	respObj, err := JSONRPCRequest(rpcURL, "eth_estimateGas", []interface{}{reqData}, 1)
	if err != nil {
		return false, err
	}

	// Parse the result to uint64
	var gasEstimationStr string
	var ok bool
	if gasEstimationStr, ok = respObj.Result.(string); !ok {
		return false, errors.New("gas estimation is not a string")
	}
	gasEstimation, err := hexutil.DecodeUint64(gasEstimationStr)
	if err != nil {
		return false, err
	}

	var minAcceptableGas uint64 = gasEstimation * 7 / 10

	// Check if the gas estimation is within the limits
	if minAcceptableGas > tx.Gas() {
		return false, fmt.Errorf("gas limit is likely too low: estimation %d, limit %d", minAcceptableGas, tx.Gas())
	}

	return true, nil
}
