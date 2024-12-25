package controllers

import (
	"net/http"
	"openeyes/services"

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
	userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
	
	query := ctx.Query("")
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
