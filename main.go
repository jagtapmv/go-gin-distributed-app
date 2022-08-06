package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal(file, &recipes)
}

func NewRecipeHandler(ctx *gin.Context) {
	var recipe Recipe

	if err := ctx.ShouldBindJSON(&recipe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"Error": err.Error(),
		})
		return
	}

	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	ctx.JSON(http.StatusOK, recipe)
}

func ListRecipeHendler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, recipes)
}

func main() {
	router := gin.Default()

	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipeHendler)

	router.Run()
}
