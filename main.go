package main

import (
	"encoding/json"
	"example/hello/util"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/NovaSubDAO/nova-sdk/go/pkg/sdk"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

const REFERRAL_CODE = 1337
const ETH_URL = "https://rpc.ankr.com/eth"
const ETH_CHAINID = 1

type (
	PositionPostParams struct {
		Address string `json:"address" validate:"required,address"`
	}

	SlippagePostParams struct {
		Amount string `json:"amount" validate:"required"`
	}

	CreateDepositTransactionParams struct {
		From   string `json:"from" validate:"required,address"`
		Token  string `json:"token" validate:"required,address"`
		Amount string `json:"amount" validate:"required"`
	}

	CreateWithdrawTransactionParams struct {
		From   string `json:"from" validate:"required,address"`
		Token  string `json:"token" validate:"required,address"`
		Amount string `json:"amount" validate:"required"`
	}

	ErrorResponse struct {
		Error       bool
		FailedField string
		Tag         string
		Value       interface{}
	}

	XValidator struct {
		validator *validator.Validate
	}

	GlobalErrorHandlerResp struct {
		Message string `json:"message"`
	}
)

var validate = validator.New()

func (v XValidator) Validate(data interface{}) []ErrorResponse {
	validationErrors := []ErrorResponse{}

	errs := validate.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			// In this case data object is actually holding the User struct
			var elem ErrorResponse

			elem.FailedField = err.Field() // Export struct field name
			elem.Tag = err.Tag()           // Export struct tag
			elem.Value = err.Value()       // Export field value
			elem.Error = true

			validationErrors = append(validationErrors, elem)
		}
	}

	return validationErrors
}

func MakeErrors(errs []ErrorResponse) *fiber.Error {
	errMsgs := make([]string, 0)

	for _, err := range errs {
		errMsgs = append(errMsgs, fmt.Sprintf(
			"[%s]: '%v' | Needs to implement '%s'",
			err.FailedField,
			err.Value,
			err.Tag,
		))
	}

	return &fiber.Error{
		Code:    fiber.StatusBadRequest,
		Message: strings.Join(errMsgs, " and "),
	}
}

func main() {
	myValidator := &XValidator{validator: validate}
	myValidator.validator.RegisterValidation("address", func(fl validator.FieldLevel) bool {
		return util.IsValidAddress(fl.Field().String())
	})

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: err.Error(),
			})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowPrivateNetwork: true,
	}))

	client, err := sdk.NewNovaSDK(ETH_URL, ETH_CHAINID)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Connected to chainid '%d' with RPC %s\n", ETH_CHAINID, ETH_URL)

	app.Get("/main/price", func(c fiber.Ctx) error {
		number, err := client.SdkDomain.GetPrice()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}
		return c.JSON(fiber.Map{
			"price": util.ToDecimal(number, 18).String(),
		})
	})

	app.Post("/main/position", func(c fiber.Ctx) error {
		params := new(PositionPostParams)
		if err := json.Unmarshal(c.Body(), params); err != nil {
			log.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid request body",
			})
		}

		// Validation
		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		addr := common.HexToAddress(params.Address)

		number, err := client.SdkDomain.GetPosition(addr)
		if err != nil {
			c.SendStatus(500)
			return c.SendString(err.Error())
		}
		return c.JSON(fiber.Map{
			"position": util.ToDecimal(number, 18).String(),
		})
	})

	app.Post("/main/slippage", func(c fiber.Ctx) error {
		params := new(SlippagePostParams)
		if err := json.Unmarshal(c.Body(), params); err != nil {
			log.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid request body",
			})
		}

		// Validation
		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		amount, err := strconv.ParseUint(params.Amount, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid amount",
			})
		}

		slippage, err := client.SdkDomain.GetSlippage(big.NewInt(int64(amount)))
		if err != nil {
			c.SendStatus(500)
			return c.SendString(err.Error())
		}

		return c.JSON(fiber.Map{
			"slippage": util.ToDecimal(slippage, 18).String(),
		})
	})

	//CreateDepositTransaction(common.Address, common.Address, *big.Int, *big.Int) (string, error)
	app.Post("/main/createDepositTx", func(c fiber.Ctx) error {
		params := new(CreateDepositTransactionParams)
		log.Println(string(c.Body()))
		if err := json.Unmarshal(c.Body(), params); err != nil {
			log.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid request body",
			})
		}
		// Validation
		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		from := common.HexToAddress(params.From)
		token := common.HexToAddress(params.Token)

		amount, err := strconv.ParseUint(params.Amount, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid amount",
			})
		}

		calldata, err := client.SdkDomain.CreateDepositTransaction(
			from,
			token,
			big.NewInt(int64(amount)),
			big.NewInt(REFERRAL_CODE),
		)

		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"calldata": calldata,
		})

	})

	//CreateWithdrawTransaction(common.Address, common.Address, *big.Int, *big.Int) (string, error)
	app.Post("/main/createWithdrawTx", func(c fiber.Ctx) error {
		params := new(CreateWithdrawTransactionParams)
		log.Println(string(c.Body()))
		if err := json.Unmarshal(c.Body(), params); err != nil {
			log.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid request body",
			})
		}
		// Validation
		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		amount, err := strconv.ParseUint(params.Amount, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid amount",
			})
		}

		from := common.HexToAddress(params.From)
		token := common.HexToAddress(params.Token)

		calldata, err := client.SdkDomain.CreateWithdrawTransaction(
			from,
			token,
			big.NewInt(int64(amount)),
			big.NewInt(REFERRAL_CODE),
		)

		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"calldata": calldata,
		})
	})

	app.Static("/", "./public")

	log.Fatal(app.Listen(":8000"))
}
