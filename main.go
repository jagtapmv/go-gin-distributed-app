package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

var (
	recipes    []Recipe
	ctx        context.Context
	er         error
	client     *mongo.Client
	collection *mongo.Collection
)

func init() {
	ctx = context.Background()
	client, er = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	//To make connection we have to pass MONGO_URI="mongodb://admin:passwd@localhost:27017" as env var to go run command as:
	//MONGO_URI="mongodb://admin:passwd@localhost:27017" go run main.go
	if er = client.Ping(context.TODO(), readpref.Primary()); er != nil {
		log.Fatal(er)
	}
	log.Println("Connected to MongoDB")
	collection = client.Database(os.Getenv("MONGO_DB")).Collection("recipes")
}

func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": err.Error(),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, er = collection.InsertOne(ctx, recipe)
	if er != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error creating the new recipe",
		})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

func ListRecipeHendler(c *gin.Context) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}

	c.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)

	_, er = collection.UpdateOne(ctx, bson.M{
		"_id": objectId,
	},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "name", Value: recipe.Name},
			{Key: "tags", Value: recipe.Tags},
			{Key: "instructions", Value: recipe.Instructions},
			{Key: "ingredients", Value: recipe.Ingredients},
		}}})
	if er != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": er.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "The recipe is updated successfully!",
	})
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectId, _ := primitive.ObjectIDFromHex(id)

	_, er = collection.DeleteOne(ctx, bson.M{
		"_id": objectId,
	})
	if er != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": er.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Success": fmt.Sprintf("The recipe with id: %s is deleted", id),
	})
}

func SearchRecipeHandler(c *gin.Context) {
	id := c.Query("id")

	objectId, _ := primitive.ObjectIDFromHex(id)

	result := collection.FindOne(ctx, bson.M{
		"_id": objectId,
	})

	var finalResult Recipe
	er = result.Decode(&finalResult)

	if er != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": er.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Searched Recipe": finalResult,
	})
}

func main() {
	router := gin.Default()

	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipeHendler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipeHandler)

	router.Run()
}
