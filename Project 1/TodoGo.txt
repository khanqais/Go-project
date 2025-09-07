package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type Todo struct {
	ID        int    `json:"id"`
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
}

func todo_main() {
	app := fiber.New()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error Loading Env file")
	}
	PORT := os.Getenv("PORT")
	todos := []Todo{}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"msg": "Hello World"})
	})

	app.Post("/api/todo", func(c *fiber.Ctx) error {
		fmt.Println("📥 POST /api/todo called")
		todo := &Todo{}
		if err := c.BodyParser(todo); err != nil {
			fmt.Println("❌ Body parsing failed:", err)
			return err
		}
		if todo.Body == "" {
			fmt.Println("❌ Empty todo body")
			return c.Status(400).JSON(fiber.Map{"Error": "Todo has a Error"})
		}
		todo.ID = len(todos) + 1
		todos = append(todos, *todo)
		fmt.Printf("✅ Todo added: %+v\n", *todo)
		fmt.Printf("📝 Total todos now: %d\n", len(todos))
		fmt.Printf("🗂️ All todos: %+v\n", todos)
		return c.Status(201).JSON(todo)
	})

	app.Get("/api/todos", func(c *fiber.Ctx) error {
		fmt.Println("📤 GET /api/todos called")
		fmt.Printf("📊 Returning %d todos: %+v\n", len(todos), todos)
		return c.Status(200).JSON(todos)
	})

	app.Patch("/api/todo/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		fmt.Printf("🔧 PATCH /api/todo/%s called\n", id)

		for i, todo := range todos {
			if fmt.Sprint(todo.ID) == id {
				todos[i].Completed = true
				fmt.Printf("✅ Todo %s marked as completed\n", id)
				return c.Status(200).JSON(todos[i])
			}
		}
		fmt.Printf("❌ Todo with ID %s not found\n", id)
		return c.Status(400).JSON(fiber.Map{"error": "Todo Not found"})
	})
	app.Delete("/api/todo/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		for i, todo := range todos {
			if fmt.Sprint(todo.ID) == id {
				todos = append(todos[:i], todos[i+1:]...)
				return c.Status(200).JSON(fiber.Map{"success": true})
			}
		}
		return c.Status(400).JSON(fiber.Map{"error": "Todo not found"})
	})

	log.Fatal(app.Listen(":" + PORT))

}
