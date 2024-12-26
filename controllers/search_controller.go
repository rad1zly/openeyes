package controllers

import (
	"net/http"
	"openeyes/services"
	"openeyes/handlers"

	"github.com/gin-gonic/gin"
)

type SearchController struct {
	searchService *services.SearchService
}

func NewSearchController(searchService *services.SearchService) *SearchController {
	return &SearchController{
		searchService: searchService,
	}
}

func (c *SearchController) Search(ctx *gin.Context) {
	user, err := handlers.authenticate(ctx)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter is required"})
		return
	}

	results, err := c.searchService.Search(query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, results)
}
