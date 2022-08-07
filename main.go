package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	handler "github.com/jagtapmv/go-gin-distributed-app/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var recipeHandler *handler.RecipeHandler

func init() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	//To make connection we have to pass MONGO_URI="mongodb://admin:passwd@localhost:27017" as env var to go run command as:
	//MONGO_URI="mongodb://admin:passwd@localhost:27017" go run main.go
	if er := client.Ping(context.TODO(), readpref.Primary()); er != nil {
		log.Fatal(er)
	}
	log.Println("Connected to MongoDB")
	collection := client.Database(os.Getenv("MONGO_DB")).Collection("recipes")

	//Passing context and collectiobn to handler
	recipeHandler = handler.NewRecipeHandler(ctx, collection)
}

func main() {
	router := gin.Default()

	router.POST("/recipes", recipeHandler.NewRecipeHandler)
	router.GET("/recipes", recipeHandler.ListRecipeHendler)
	router.PUT("/recipes/:id", recipeHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", recipeHandler.DeleteRecipeHandler)
	router.GET("/recipes/search", recipeHandler.SearchRecipeHandler)

	router.Run()
}
