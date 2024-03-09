package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// JSONRPCRequest represents a JSON-RPC request payload
type JSONRPCRequestData struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC response payload
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func JSONRPCRequest(rpcURL string, method string, params interface{}, id int) (JSONRPCResponse, error) {
	requestPayload := JSONRPCRequestData{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	log.Printf("Sending JSON-RPC request to %s: %+v\n", rpcURL, params)

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return JSONRPCResponse{}, err
	}

	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return JSONRPCResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return JSONRPCResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return JSONRPCResponse{}, err
	}

	var respObj JSONRPCResponse
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return JSONRPCResponse{}, err
	}

	if respObj.Error != nil {
		return JSONRPCResponse{}, fmt.Errorf("JSON RPC Error: %v", respObj.Error)
	}

	return respObj, nil
}
