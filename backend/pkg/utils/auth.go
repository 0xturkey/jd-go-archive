package utils

import (
	"errors"
	"strings"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spruceid/siwe-go"
)

type VerifySignatureParams struct {
	Signature       string
	Message         string
	ExpectedAddress string
}

func VerifySignature(params *VerifySignatureParams) error {
	var err error
	var siweMessage *siwe.Message
	siweMessage, err = siwe.ParseMessage(params.Message)
	if err != nil {
		return err
	}

	var publicKey *ecdsa.PublicKey
	publicKey, err = siweMessage.Verify(params.Signature, nil, nil, nil)
	if err != nil {
		return err
	}

	address := crypto.PubkeyToAddress(*publicKey).Hex()
	if !strings.EqualFold(address, params.ExpectedAddress) {
		return errors.New("PrimaryAddress does not match signature")
	}

	return nil
}

func NormalizeAddress(address string) (string, error) {
	if !common.IsHexAddress(address) {
		return "", errors.New("invalid address")
	}

	normalized := common.HexToAddress(address).Hex()

	return normalized, nil
}
