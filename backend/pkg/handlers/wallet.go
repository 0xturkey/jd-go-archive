package handlers

import (
	"fmt"

	"github.com/0xturkey/jd-go/database"
	"github.com/0xturkey/jd-go/model"
	"github.com/0xturkey/jd-go/pkg/turnkey"
	"github.com/0xturkey/jd-go/pkg/types"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	tkWallet "github.com/tkhq/go-sdk/pkg/api/client/wallets"
	tkModels "github.com/tkhq/go-sdk/pkg/api/models"
)

// FindOrCreateWallet looks up a wallet by address, or creates a new one if not found.
/*
func FindOrCreateWallet(address string, userID uuid.UUID) (*model.Wallet, error) {
	db := database.DB
	wallet := &model.Wallet{}
	normalizedAddress, err := utils.NormalizeAddress(address)
	if err != nil {
		return nil, err
	}

	err = db.Where("address = ? AND user_refer = ?", normalizedAddress, userID).First(wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Wallet not found, create a new wallet
			wallet.Address = normalizedAddress
			wallet.UserRefer = userID
			err = db.Create(wallet).Error
			if err != nil {
				return nil, err
			}
		} else {
			// Some other error occurred
			return nil, err
		}
	}
	// Wallet found or created successfully
	return wallet, nil
}
*/

/*
func CreateWallet(c *fiber.Ctx) error {
	db := database.DB
	json := new(model.Wallet)
	if err := c.BodyParser(json); err != nil {
		return c.JSON(fiber.Map{
			"code":    400,
			"message": "Invalid JSON",
		})
	}
	userID := c.Locals("user_id").(uuid.UUID)
	newWallet := model.Wallet{
		UserRefer: userID,
		Address:   json.Address,
	}
	err := db.Create(&newWallet).Error
	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "sucess",
	})
}
*/

func GetWallet(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(uuid.UUID)

	var wallet model.Wallet
	err := db.Preload("User").Where("wallets.user_refer = ?", userID).First(&wallet).Error
	if err != nil {
		// Handle error, e.g., return an internal server error status
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"message": "Could not retrieve wallets",
		})
	}

	p := tkWallet.NewGetWalletAccountsParams().WithBody(&tkModels.GetWalletAccountsRequest{
		OrganizationID: &wallet.User.SuborganizationID,
		WalletID:       &wallet.TurnkeyID,
	})

	response, err := turnkey.Client.Client.Wallets.GetWalletAccounts(p, turnkey.Client.GetAuthenticator())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if response.Code() != 200 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
		})
	}

	// Return the wallets associated with the user
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "success",
		"data":    response.Payload.Accounts,
	})
}

/*
	var req types.ExportRequest
	err = ctx.BindJSON(&req)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	bodyBytes, err := turnkey.Client.ForwardSignedActivity(req.SignedExportRequest.Url, req.SignedExportRequest.Body, req.SignedExportRequest.Stamp)
	if err != nil {
		err = errors.Wrap(err, "error while forwarding signed EXPORT_WALLET activity")
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	exportBundle := gjson.Get(string(bodyBytes), "activity.result.exportWalletResult.exportBundle").String()

	ctx.JSON(http.StatusOK, exportBundle)
*/

func ExportWallet(c *fiber.Ctx) error {
	json := new(types.ExportRequest)
	if err := c.BodyParser(json); err != nil {
		return c.JSON(fiber.Map{
			"code":    400,
			"message": "Invalid Request Data",
		})
	}

	fmt.Println("handling export wallet request")

	bodyBytes, err := turnkey.Client.ForwardSignedActivity(json.SignedExportRequest.Url, json.SignedExportRequest.Body, json.SignedExportRequest.Stamp)
	if err != nil {
		err = errors.Wrap(err, "error while forwarding signed EXPORT_WALLET activity")
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	exportBundle := gjson.Get(string(bodyBytes), "activity.result.exportWalletResult.exportBundle").String()

	return c.JSON(fiber.Map{
		"bundle": exportBundle,
	})
}
