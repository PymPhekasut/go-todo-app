package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("Hello Worlds")

	if os.Getenv("ENV") != "production" {
		// load the .env file if not in the production
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(("Connected to MongoDB"))
	collection = client.Database("golang_db").Collection("todos")

	app := fiber.New()

	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: "http://localhost:5173",
	// 	AllowHeaders: "Origin, Content-type, Accept",
	// }))

	app.Get("/api/todos", getTodos)
	app.Get("/api/todos/:id", getTodoById)
	app.Post("/api/todos", createTodos)
	app.Patch("/api/todos/:id", updateTodos)
	app.Delete("/api/todos/:id", deleteTodos)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/dist")
	}
	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo

	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}
	return c.JSON(todos)
}

func createTodos(c *fiber.Ctx) error {
	todo := new(Todo) // {id:0, completed: true, body: ""}

	// pass request body
	if err := c.BodyParser((todo)); err != nil {
		return err
	}
	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo bodt cannot be empty"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)
	return c.Status(201).JSON(todo)

}

func getTodoById(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}
	var todo Todo

	filter := bson.M{"_id": objectId}
	err = collection.FindOne(context.Background(), filter).Decode(&todo)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Not found todo"})
	}
	return c.Status(200).JSON(todo)

}

func updateTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}
	// Parse the request body into a new Todo struct
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON((fiber.Map{"success": true}))

}

// func updateTodos(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	objectID, err := primitive.ObjectIDFromHex(id)
// 	if err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
// 	}
// 	// Parse the request body into a new Todo struct
// 	todo := new(Todo) // {id:0, completed: true, body: ""}
// 	// pass request body
// 	if err := c.BodyParser((todo)); err != nil {
// 		return err
// 	}

// 	filter := bson.M{"_id": objectID}
// 	update := bson.M{"$set": bson.M{"body": todo.Body, "completed": todo.Completed}}

// 	result, err := collection.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		return err
// 	}

// 	// Check if any document was updated
// 	if result.MatchedCount == 0 {
// 		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
// 	}
// 	return c.Status(200).JSON((fiber.Map{"success": "Updated successfully"}))

// }

func deleteTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	filter := bson.M{"_id": objectId}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return c.Status(200).JSON((fiber.Map{"msg": "Deleted successfully."}))
}
