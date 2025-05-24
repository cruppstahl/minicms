package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Post represents a simple post structure
type Post struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Mock data for posts
var posts = []Post{
	{ID: "1", Title: "First Post", Body: "This is the body of the first post."},
	{ID: "2", Title: "Second Post", Body: "This is the body of the second post."},
}

// GetPosts handles requests to retrieve all posts
func GetPosts(c *gin.Context) {
	c.JSON(http.StatusOK, posts)
}

// GetPostByID handles requests to retrieve a specific post by ID
func GetPostByID(c *gin.Context) {
	id := c.Param("post-id")
	for _, post := range posts {
		if post.ID == id {
			c.JSON(http.StatusOK, post)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "post not found"})
}
