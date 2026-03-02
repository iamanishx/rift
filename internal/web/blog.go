package web

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamanishx/xserve/internal/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreatePost(c *gin.Context) {
	uid := c.GetString("user_id")

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		IsPublic bool  `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	slug := generateSlug(input.Title)
	exists, _ := db.GetPostBySlug(uid, slug)
	if exists != nil {
		slug = slug + "-" + time.Now().Format("20060102150405")
	}

	post := &db.Post{
		UserID:   uid,
		Slug:     slug,
		Title:    input.Title,
		Content:  input.Content,
		IsPublic: input.IsPublic,
	}

	if err := db.CreatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func GetPosts(c *gin.Context) {
	uid := c.GetString("user_id")
	posts, err := db.GetUserPosts(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}
	c.JSON(http.StatusOK, posts)
}

func GetPost(c *gin.Context) {
	uid := c.GetString("user_id")
	slug := c.Param("slug")

	post, err := db.GetPostBySlug(uid, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func UpdatePost(c *gin.Context) {
	uid := c.GetString("user_id")
	slug := c.Param("slug")

	post, err := db.GetPostBySlug(uid, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		IsPublic *bool `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if input.Title != "" {
		newSlug := generateSlug(input.Title)
		if newSlug != post.Slug {
			exists, _ := db.GetPostBySlug(uid, newSlug)
			if exists == nil {
				post.Slug = newSlug
			}
		}
		post.Title = input.Title
	}

	if input.Content != "" {
		post.Content = input.Content
	}

	if input.IsPublic != nil {
		post.IsPublic = *input.IsPublic
	}

	if err := db.UpdatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func DeletePost(c *gin.Context) {
	uid := c.GetString("user_id")
	slug := c.Param("slug")

	post, err := db.GetPostBySlug(uid, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	if err := db.DeletePost(post.ID.Hex()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}

func ToggleVisibility(c *gin.Context) {
	uid := c.GetString("user_id")
	slug := c.Param("slug")

	post, err := db.GetPostBySlug(uid, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	updated, err := db.TogglePostVisibility(post.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle visibility"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func GetPublicPosts(c *gin.Context) {
	username := c.Param("username")

	user, err := db.GetUser(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	posts, err := db.GetPublicPosts(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"posts": posts,
	})
}

func GetPublicPost(c *gin.Context) {
	username := c.Param("username")
	slug := c.Param("slug")

	user, err := db.GetUser(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	post, err := db.GetPublicPostBySlug(user.ID, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"post": post,
	})
}

func generateSlug(title string) string {
	reg := regexp.MustCompile(`[^\w\s-]`)
	slug := reg.ReplaceAllString(title, "")
	slug = strings.TrimSpace(slug)
	slug = strings.ToLower(slug)
	slug = regexp.MustCompile(`[\s_]+`).ReplaceAllString(slug, "-")
	slug = filepath.Base(slug)
	if slug == "" || slug == "-" {
		slug = "untitled"
	}
	return slug
}


