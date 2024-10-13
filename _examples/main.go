package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/satjan/context"
	"log"
)

type CreateDataReq struct {
	Name string `json:"name" validate:"required,min=3,max=5"`
	Code string `json:"code" validate:"required"`
}

func main() {
	app := fiber.New()

	app.Post("/", func(ctx *fiber.Ctx) error {
		req := new(CreateDataReq)
		if err := context.Parse(ctx, req); err != nil {
			return nil
		}

		return context.Response(ctx, req, nil, "")
	})

	err := app.Listen(":3000")
	if err != nil {
		log.Fatalln(err)
	}
}
