package main

import (
	"log"

	"github.com/NovaSubDAO/nova-sdk/go/pkg/sdk"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	// FIXME: where do I put RPC?
	// NOTE: .........................vault address
	client, cleanup := sdk.NewNovaSDK("0x1234123412341234123412341234123412341234")
	defer cleanup()

	app.Get("/price", func(c fiber.Ctx) error {
		// get price from the sdk
		price, err := client.GetPrice()
		if err != nil {
			c.SendStatus(500)
			c.SendString(err.Error())
		}

		return c.SendString(price.String())
	})

	log.Fatal(app.Listen(":8000"))
}
