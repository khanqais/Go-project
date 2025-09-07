package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo_mongoo struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Completed bool               `json:"completed" bson:"completed"`
	Body      string             `json:"body" bson:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("Starting Todo API...")

	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get MongoDB URI
	mongoURL := os.Getenv("MONGO_URI")
	if mongoURL == "" {
		log.Fatal("MONGO_URI not found in environment variables")
	}

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(mongoURL)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	// Ensure disconnection on exit
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal("Error disconnecting from MongoDB:", err)
		}
	}()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB:", err)
	}

	fmt.Println("Connected to MongoDB successfully!")

	// Set up collection
	collection = client.Database("GO").Collection("todos")

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"message": "Todo API is running!",
			"version": "1.0.0",
		})
	})

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	// Get port
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}

	fmt.Printf("Server starting on port %s...\n", PORT)
	log.Fatal(app.Listen("0.0.0.0:" + PORT))
}

// Get all todos
func getTodos(c *fiber.Ctx) error {
	fmt.Println(" getTodos function called!")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("About to query MongoDB...")

	var todos []Todo_mongoo
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Printf("MongoDB query failed: %v\n", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch todos",
		})
	}
	defer cursor.Close(ctx)

	fmt.Println("MongoDB query successful")

	if err = cursor.All(ctx, &todos); err != nil {
		fmt.Printf(" Cursor decode failed: %v\n", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to decode todos",
		})
	}

	fmt.Printf("Found %d todos\n", len(todos))

	if todos == nil {
		todos = []Todo_mongoo{}
	}

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    todos,
		"count":   len(todos),
	})
}

// Create a new todo
func createTodo(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var todo Todo_mongoo

	// Parse request body
	if err := c.BodyParser(&todo); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Todo body is required",
		})
	}

	// Set default values
	todo.ID = primitive.NewObjectID()
	if todo.Completed == false {
		todo.Completed = false // Explicit default
	}

	// Insert into database
	result, err := collection.InsertOne(ctx, todo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create todo",
		})
	}

	// Set the ID from the insert result
	todo.ID = result.InsertedID.(primitive.ObjectID)

	fmt.Printf("Todo created: %+v\n", todo)

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    todo,
	})
}

// Update todo completion status
func updateTodo(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get ID from URL params
	idParam := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid todo ID",
		})
	}

	// Parse request body for updates
	var updateData bson.M
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create update document
	update := bson.M{"$set": updateData}

	// Update the document
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		update,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update todo",
		})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}

	// Fetch and return updated todo
	var updatedTodo Todo_mongoo
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&updatedTodo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch updated todo",
		})
	}

	fmt.Printf("Todo updated: %+v\n", updatedTodo)

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    updatedTodo,
	})
}

// Delete a todo
func deleteTodo(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get ID from URL params
	idParam := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid todo ID",
		})
	}

	// Delete the document
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete todo",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}

	fmt.Printf("Todo deleted with ID: %s\n", idParam)

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"message": "Todo deleted successfully",
	})
}
