package utils

import (
	"net/url"
	"strings"
	"time"

	"github.com/0xturkey/jd-go/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func isEthereumAddress(fl validator.FieldLevel) bool {
	address := fl.Field().String()
	return common.IsHexAddress(address)
}

// isHTTPSURL checks if the field is a valid HTTPS URL.
func isHTTPSURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	parsedURL, err := url.Parse(urlStr)
	return err == nil && parsedURL.Scheme == "https"
}

// isHexPrefixed checks if the field is a hex string prefixed with "0x".
func isHexString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return strings.HasPrefix(value, "0x") && isHex(value[2:])
}

// isHex checks if a string is a valid hexadecimal.
func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func isValidUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

const timeLayout = time.RFC3339 // or "2006-01-02T15:04:05Z07:00" for example

func isRFC3339Time(fl validator.FieldLevel) bool {
	_, err := time.Parse(timeLayout, fl.Field().String())
	return err == nil
}

// validateTaskType is the custom validation function for TaskType
func validateTaskType(fl validator.FieldLevel) bool {
	_, err := GetTaskHandler(model.TaskType(fl.Field().String()))
	return err == nil
}

func NewValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterValidation("isEthAddress", isEthereumAddress)
	validate.RegisterValidation("isHTTPSURL", isHTTPSURL)
	validate.RegisterValidation("isHexString", isHexString)
	validate.RegisterValidation("isValidUUID", isValidUUID)
	validate.RegisterValidation("isRFC3339Time", isRFC3339Time)
	validate.RegisterValidation("validateTaskType", validateTaskType)
	return validate
}
