package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

type Todo struct {
	ID        int    `json:"id"`
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
}

func main() {
	app := fiber.New()
	todos := []Todo{}
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"msg": "Hello World"})
	})
	app.Post("/api/todo", func(c *fiber.Ctx) error {
		todo := &Todo{}
		if err := c.BodyParser(todo); err != nil {
			return err
		}
		if todo.Body == "" {
			return c.Status(400).JSON(fiber.Map{"Error": "Todo has a Error"})
		}
		todo.ID = len(todos) + 1
		todos = append(todos, *todo)
		fmt.Printf("Todos: %+v\n", todos)
		return c.Status(201).JSON(todo)
	})

	log.Fatal(app.Listen(":3000"))
}
