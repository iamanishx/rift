package main

import (
	"html/template"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/iamanishx/xserve/internal/auth"
	"github.com/iamanishx/xserve/internal/db"
	"github.com/iamanishx/xserve/internal/engine"
	"github.com/iamanishx/xserve/internal/web"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	if err := db.Connect(os.Getenv("MONGO_URI")); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	engine.InitStorage()

	r := gin.Default()

	funcs := template.FuncMap{
		"now":          web.Now,
		"calcReadTime": web.CalcReadTime,
		"truncate":     web.Truncate,
		"renderMarkdown": web.RenderMarkdown,
	}
	r.SetFuncMap(funcs)

	r.LoadHTMLGlob("internal/web/templates/*")

	auth.Setup(r)

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})
	r.GET("/auth/google", web.AuthLogin)
	r.GET("/auth/google/callback", web.AuthCallback)

	r.GET("/:username", web.BlogIndex)
	r.GET("/:username/:slug", web.BlogPost)

	authorized := r.Group("/")
	authorized.Use(auth.AuthMiddleware())
	{
		authorized.GET("/dashboard", web.Dashboard)
		authorized.POST("/upload", web.Upload)
		authorized.GET("/preview/:slug", web.BlogPostPreview)

		authorized.POST("/api/posts", web.CreatePost)
		authorized.GET("/api/posts", web.GetPosts)
		authorized.GET("/api/posts/:slug", web.GetPost)
		authorized.PUT("/api/posts/:slug", web.UpdatePost)
		authorized.DELETE("/api/posts/:slug", web.DeletePost)
		authorized.PATCH("/api/posts/:slug/visibility", web.ToggleVisibility)
	}

	r.Static("/sites", "./data/sites")

	log.Println("Server running on :8080")
	r.Run(":8080")
}
