package handlers

import (
	"fmt"
	"time"

	"github.com/0xturkey/jd-go/database"
	"github.com/0xturkey/jd-go/model"
	"github.com/0xturkey/jd-go/pkg/turnkey"
	"github.com/0xturkey/jd-go/pkg/types"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	tkUsers "github.com/tkhq/go-sdk/pkg/api/client/users"
	tkModels "github.com/tkhq/go-sdk/pkg/api/models"
	"gorm.io/gorm"
)

var JwtSecretKey = []byte("your_secret_key")

type User model.User

func GenerateJWT(user User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["user_id"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // Token expires after 72 hours

	tokenString, err := token.SignedString(JwtSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUser(id uuid.UUID) (User, error) {
	db := database.DB

	user := User{}
	usrQuery := User{ID: id}
	var err = db.First(&user, &usrQuery).Error
	if err == gorm.ErrRecordNotFound {
		return User{}, err
	}
	return user, nil
}

func Login(c *fiber.Ctx) error {
	db := database.DB

	json := new(types.AuthenticationRequest)
	if err := c.BodyParser(json); err != nil {
		return c.JSON(fiber.Map{
			"code":    400,
			"message": "Invalid Request Data",
		})
	}

	status, bodyBytes, err := turnkey.Client.ForwardSignedRequest(json.SignedCreateAPIKeyRequest.Url, json.SignedCreateAPIKeyRequest.Body, json.SignedCreateAPIKeyRequest.Stamp)
	if err != nil {
		err = errors.Wrap(err, "error while forwarding signed request to create a session key")
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if status != 200 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": fmt.Sprintf("expected 200 when forwarding request to create a session key. Got %d", status),
		})
	}

	subOrganizationId := gjson.Get(string(bodyBytes), "activity.organizationId").String()

	fmt.Println("sub org", subOrganizationId)

	found := User{}
	query := User{SuborganizationID: subOrganizationId}
	err = db.First(&found, &query).Error
	if err == gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	tokenString, err := GenerateJWT(found)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not generate token",
		})
	}

	return c.JSON(fiber.Map{
		"message":           "success",
		"email":             found.Email,
		"userID":            found.ID,
		"suborganizationID": found.SuborganizationID,
		"token":             tokenString,
	})
}

func CreateUser(c *fiber.Ctx) error {
	var err error
	db := database.DB

	type CreateUserReq struct {
		Email       string `json:"email"`
		Attestation types.Attestation
		Challenge   string
	}
	json := new(CreateUserReq)
	if err = c.BodyParser(json); err != nil {
		return c.JSON(fiber.Map{
			"code":    400,
			"message": "Invalid Request Data",
		})
	}

	found := User{}
	query := User{Email: json.Email}
	err = db.First(&found, &query).Error
	if err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User already exists",
		})
	}

	subOrgResult, err := turnkey.Client.CreateUserSubOrganization(json.Email, json.Attestation, json.Challenge)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
		})
	}

	p := tkUsers.NewGetUsersParams().WithBody(&tkModels.GetUsersRequest{
		OrganizationID: &subOrgResult.SubOrganizationID,
	})
	response, err := turnkey.Client.Client.Users.GetUsers(p, turnkey.Client.GetAuthenticator())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	new := User{
		Email:             json.Email,
		SuborganizationID: subOrgResult.SubOrganizationID,
		TkID:              *response.Payload.Users[0].UserID,
		ID:                uuid.New(),
	}
	db.Create(&new)

	wallet := model.Wallet{
		ID:        uuid.New(),
		TurnkeyID: subOrgResult.WalletID,
		UserRefer: new.ID,
	}
	db.Create(&wallet)

	return c.JSON(fiber.Map{
		"message":           "success",
		"userID":            new.ID,
		"suborganizationID": new.SuborganizationID,
		"tkID":              new.TkID,
	})
}

func UserExists(c *fiber.Ctx) error {
	db := database.DB

	email := c.Params("email")

	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email parameter is required",
		})
	}

	found := User{}
	query := User{Email: email}
	err := db.First(&found, &query).Error
	if err == gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusNoContent).JSON(fiber.Map{})
	}

	return c.JSON(fiber.Map{
		"message":           "User exists",
		"userID":            found.ID,
		"suborganizationID": found.SuborganizationID,
		"tkID":              found.TkID,
	})
}

func GetUserInfo(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	user, err := GetUser(userID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}

	return c.JSON(user)
}
