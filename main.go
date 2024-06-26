package main

import (
	"encoding/json"
	"example/hello/util"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/NovaSubDAO/nova-sdk/go/pkg/constants"
	"github.com/NovaSubDAO/nova-sdk/go/pkg/sdk"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

const REFERRAL_CODE = 1337
const ETH_URL = "https://rpc.ankr.com/eth"
const ETH_CHAINID = 1
const OPT_URL = "https://rpc.ankr.com/optimism"
const OPT_CHAINID = 10

type (
	PositionPostParams struct {
		// Stablecoin constants.Stablecoin `json:"stablecoin" validate:"required"`
		Address string `json:"address" validate:"required,address"`
	}

	SlippagePostParams struct {
		InputToken constants.Stablecoin `json:"inputToken" validate:"required"`
		Amount     string               `json:"amount" validate:"required"`
	}

	CreateDepositTransactionParams struct {
		From   string               `json:"from" validate:"required,address"`
		Token  constants.Stablecoin `json:"token" validate:"required"`
		Amount string               `json:"amount" validate:"required"`
	}

	CreateWithdrawTransactionParams struct {
		From   string               `json:"from" validate:"required,address"`
		Token  constants.Stablecoin `json:"token" validate:"required"`
		Amount string               `json:"amount" validate:"required"`
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

	ethClient, err := sdk.NewNovaSDK(ETH_URL, ETH_CHAINID)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Connected to chainid '%d' with RPC %s\n", ETH_CHAINID, ETH_URL)

	optClient, err := sdk.NewNovaSDK(OPT_URL, OPT_CHAINID)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Connected to chainid '%d' with RPC %s\n", OPT_CHAINID, OPT_URL)

	app.Get("/main/canonicalPrice", func(c fiber.Ctx) error {
		// NOTE: Get the canonical price from Mainnet.
		price, err := ethClient.GetSDaiPrice()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}
		return c.JSON(fiber.Map{
			"price": util.ToDecimal(price, 18).String(),
		})
	})

	app.Get("/opt/canonicalPrice", func(c fiber.Ctx) error {
		// NOTE: Get the canonical price from Optimism.
		price, err := optClient.GetSDaiPrice()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}
		return c.JSON(fiber.Map{
			"price": util.ToDecimal(price, 18).String(),
		})
	})

	app.Get("/main/price", func(c fiber.Ctx) error {
		number, err := ethClient.GetPrice(constants.DAI)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}
		return c.JSON(fiber.Map{
			"price": util.ToDecimal(number, constants.StablecoinDetails[ETH_CHAINID][constants.DAI].Decimals).String(),
		})
	})

	app.Get("/opt/price", func(c fiber.Ctx) error {
		number, err := optClient.GetPrice(constants.USDC)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}
		return c.JSON(fiber.Map{
			"price": util.ToDecimal(number, constants.StablecoinDetails[OPT_CHAINID][constants.USDC].Decimals).String(),
		})
	})

	app.Get("/main/supportedStablecoins", func(c fiber.Ctx) error {
		stablecoins, err := ethClient.GetSupportedStablecoins()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		stablecoinArray := make([]fiber.Map, 0, 10)

		for _, sc := range stablecoins {
			stablecoinArray = append(stablecoinArray, fiber.Map{
				"symbol":   string(sc),
				"address":  constants.StablecoinDetails[ETH_CHAINID][sc].Address,
				"decimals": constants.StablecoinDetails[ETH_CHAINID][sc].Decimals,
			})
		}
		return c.JSON(stablecoinArray)
	})

	app.Get("/opt/supportedStablecoins", func(c fiber.Ctx) error {
		stablecoins, err := optClient.GetSupportedStablecoins()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
		}

		stablecoinArray := make([]fiber.Map, 0, 10)

		for _, sc := range stablecoins {
			stablecoinArray = append(stablecoinArray, fiber.Map{
				"symbol":   string(sc),
				"address":  constants.StablecoinDetails[OPT_CHAINID][sc].Address,
				"decimals": constants.StablecoinDetails[OPT_CHAINID][sc].Decimals,
			})
		}
		return c.JSON(stablecoinArray)
	})

	app.Post("/main/position", func(c fiber.Ctx) error {
		params := new(PositionPostParams)
		if err := json.Unmarshal(c.Body(), params); err != nil {
			log.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid request body",
			})
		}

		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		addr := common.HexToAddress(params.Address)

		number, err := ethClient.GetPosition(constants.DAI, addr)
		if err != nil {
			c.SendStatus(500)
			return c.SendString(err.Error())
		}
		return c.JSON(fiber.Map{
			"position": util.ToDecimal(number, constants.StablecoinDetails[ETH_CHAINID][constants.DAI].Decimals).String(),
		})
	})

	app.Post("/opt/position", func(c fiber.Ctx) error {
		params := new(PositionPostParams)
		if err := json.Unmarshal(c.Body(), params); err != nil {
			log.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid request body",
			})
		}

		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		addr := common.HexToAddress(params.Address)

		number, err := optClient.GetPosition(constants.USDC, addr)
		if err != nil {
			c.SendStatus(500)
			return c.SendString(err.Error())
		}
		return c.JSON(fiber.Map{
			"position": util.ToDecimal(number, constants.StablecoinDetails[OPT_CHAINID][constants.USDC].Decimals).String(),
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

		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		amount, err := strconv.ParseFloat(params.Amount, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid amount",
			})
		}

		tokenDecimals := constants.StablecoinDetails[ETH_CHAINID][params.InputToken].Decimals

		slippage, _, _, err := ethClient.GetSlippage(params.InputToken, util.ToWei(amount, tokenDecimals))
		if err != nil {
			c.SendStatus(500)
			return c.SendString(err.Error())
		}

		return c.JSON(fiber.Map{
			"slippage": slippage,
		})
	})

	app.Post("/opt/slippage", func(c fiber.Ctx) error {
		params := new(SlippagePostParams)
		if err := json.Unmarshal(c.Body(), params); err != nil {
			log.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid request body",
			})
		}

		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		amount, err := strconv.ParseFloat(params.Amount, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid amount",
			})
		}

		tokenDecimals := constants.StablecoinDetails[OPT_CHAINID][params.InputToken].Decimals

		slippage, _, _, err := optClient.GetSlippage(params.InputToken, util.ToWei(amount, tokenDecimals))
		if err != nil {
			c.SendStatus(500)
			return c.SendString(err.Error())
		}

		return c.JSON(fiber.Map{
			"slippage": slippage,
		})
	})

	app.Get("/main/vaultAddress", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"address": ethClient.GetConfig().VaultAddress,
		})
	})

	app.Get("/opt/vaultAddress", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"address": optClient.GetConfig().VaultAddress,
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

		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		from := common.HexToAddress(params.From)

		amount, success := big.NewInt(0).SetString(params.Amount, 10)
		if !success {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid amount",
			})
		}

		calldata, err := ethClient.CreateDepositTransaction(
			params.Token,
			from,
			amount,
			big.NewInt(REFERRAL_CODE),
		)

		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"calldata": calldata,
		})

	})

	app.Post("/opt/createDepositTx", func(c fiber.Ctx) error {
		params := new(CreateDepositTransactionParams)
		log.Println(string(c.Body()))
		if err := json.Unmarshal(c.Body(), params); err != nil {
			log.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid request body",
			})
		}

		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		from := common.HexToAddress(params.From)

		amount, success := big.NewInt(0).SetString(params.Amount, 10)
		if !success {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid amount",
			})
		}

		calldata, err := optClient.CreateDepositTransaction(
			params.Token,
			from,
			amount,
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

		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		amount, success := big.NewInt(0).SetString(params.Amount, 10)
		if !success {
			log.Println("Failed to parse amount")
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid amount",
			})
		}

		from := common.HexToAddress(params.From)

		calldata, err := ethClient.CreateWithdrawTransaction(
			params.Token,
			from,
			amount,
			big.NewInt(REFERRAL_CODE),
		)

		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"calldata": calldata,
		})
	})

	app.Post("/opt/createWithdrawTx", func(c fiber.Ctx) error {
		params := new(CreateWithdrawTransactionParams)
		log.Println(string(c.Body()))
		if err := json.Unmarshal(c.Body(), params); err != nil {
			log.Println(err.Error())
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid request body",
			})
		}

		if errs := myValidator.Validate(params); len(errs) > 0 && errs[0].Error {
			return MakeErrors(errs)
		}

		amount, success := big.NewInt(0).SetString(params.Amount, 10)
		if !success {
			log.Println("Failed to parse amount")
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: "Invalid amount",
			})
		}

		from := common.HexToAddress(params.From)

		calldata, err := optClient.CreateWithdrawTransaction(
			params.Token,
			from,
			amount,
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
