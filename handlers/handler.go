package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	model "github.com/jagtapmv/go-gin-distributed-app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	recipes []model.Recipe
	er      error
)

type RecipeHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipeHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipeHandler {
	return &RecipeHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

func (handler *RecipeHandler) NewRecipeHandler(c *gin.Context) {
	var recipe model.Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": err.Error(),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, er = handler.collection.InsertOne(handler.ctx, recipe)
	if er != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error creating the new recipe",
		})
		return
	}
	handler.redisClient.Del(handler.ctx, "recipes")
	c.JSON(http.StatusOK, recipe)
}

func (handler *RecipeHandler) ListRecipeHendler(c *gin.Context) {
	val, err := handler.redisClient.Get(handler.ctx, "recipes").Result()
	if err == redis.Nil {
		log.Println("Request to MongoDB")
		cur, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer cur.Close(handler.ctx)

		for cur.Next(handler.ctx) {
			var recipe model.Recipe
			cur.Decode(&recipe)
			recipes = append(recipes, recipe)
		}

		data, _ := json.Marshal(recipes)
		handler.redisClient.Set(handler.ctx, "recipes", string(data), 0)

		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": err.Error(),
		})
		return
	} else {
		log.Println("Request to Redis")
		err := json.Unmarshal([]byte(val), &recipes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, recipes)
	}
}

func (handler *RecipeHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe model.Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)

	_, er = handler.collection.UpdateOne(handler.ctx, bson.M{
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
	handler.redisClient.Del(handler.ctx, "recipes")
	c.JSON(http.StatusOK, gin.H{
		"message": "The recipe is updated successfully!",
	})
}

func (handler *RecipeHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectId, _ := primitive.ObjectIDFromHex(id)

	_, er = handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if er != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": er.Error(),
		})
		return
	}
	handler.redisClient.Del(handler.ctx, "recipes")
	c.JSON(http.StatusOK, gin.H{
		"Success": fmt.Sprintf("The recipe with id: %s is deleted", id),
	})
}

func (handler *RecipeHandler) SearchRecipeHandler(c *gin.Context) {
	id := c.Query("id")

	objectId, _ := primitive.ObjectIDFromHex(id)

	result := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	})

	var finalResult model.Recipe
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
