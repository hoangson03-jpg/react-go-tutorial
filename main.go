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
	"go.mongodb.org/mongo-driver/mongo" // gói trình điều khiển mongo để kết nối với cơ sở dữ liệu MongoDB và thực hiện các thao tác cơ bản như ping cơ sở dữ liệu.
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // Dùng dấu nháy chéo `` để tạo thành JSON object và BSON object (JavaScript & Binary) Các thẻ bson rất quan trọng để ánh xạ các trường cấu trúc Go của chúng tôi tới các trường tương ứng trong tài liệu MongoDB của chúng tôi.
	// omitempty tức là bỏ qua những giá trị rỗng hoặc ID rỗng
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("Hello World")

	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env") // Load sẽ đọc (các) tệp env của bạn và tải chúng vào ENV cho quá trình này.
		// Gọi hàm này càng gần điểm bắt đầu chương trình của bạn càng tốt (lý tưởng nhất là trong main).
		// Nếu bạn gọi Load mà không có bất kỳ đối số nào, nó sẽ mặc định tải .env trong đường dẫn hiện tại.
		// Ngoài ra, bạn có thể cho nó biết tệp nào cần tải (có thể có nhiều tệp)
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")                           // mở một tập tin và đọc đường dẫn trong đó.
	clientOptions := options.Client().ApplyURI(MONGODB_URI)           // Máy khách tạo một phiên bản ClientOptions mới.
	client, err := mongo.Connect(context.Background(), clientOptions) // Hàm `context.Background()` cung cấp bối cảnh gốc để sử dụng khi không có bối cảnh nào khác
	// 3 dòng lệnh ở trên là cần thiết để kết nối tới MongoDB

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil) // Ping gửi lệnh ping để xác minh rằng máy khách có thể kết nối với triển khai.
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MONGODB ATLAS")

	collection = client.Database("golang_db").Collection("todos")

	app := fiber.New()

	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: "http://localhost:5173",
	// 	AllowHeaders: "Origin,Content-Type,Accept",
	// }))

	app.Get("/api/todos", getTodos)
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

	cursor, err := collection.Find(context.Background(), bson.M{}) // loại M được xác định trong gói này có thể được sử dụng để xây dựng các biểu diễn BSON bằng cách sử dụng các loại Go gốc M là Map
	// cursor là con trỏ trỏ tới giá trị khi bạn thực hiện một truy vấn trong mongodb

	if err != nil {
		return err
	}

	defer cursor.Close(context.Background()) // defer là một từ khóa mà chúng ta sử dụng trong go để hoãn việc thực thi một lệnh gọi hàm cho đến khi một hàm xung quanh hoàn tất (hàm cần hoãn ở đây đso là getTodos)

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
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID) // primitive.ObjectID: Dạng nguyên thủy của ObjectID thuộc mongodb
	// Ở đây sẽ chuyển dạng từ string mà người dùng nhập vào sang dạng ObjectID

	return c.Status(201).JSON(todo)
}

func updateTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error:": "Invalid todo ID"})
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"success": "true"})
}

func deleteTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error:": "Invalid todo ID"})
	}

	filter := bson.M{"_id": objectID}
	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"deleted": "succesfully"})
}
