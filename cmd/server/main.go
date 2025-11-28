package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/iamanishx/xserve/internal/auth"
	"github.com/iamanishx/xserve/internal/db"
	"github.com/iamanishx/xserve/internal/web"
	"github.com/joho/godotenv"
	csrf "github.com/utrack/gin-csrf"
)

func main() {
	godotenv.Load()

	if err := db.Connect(os.Getenv("MONGO_URI")); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	r := gin.Default()
	r.LoadHTMLGlob("internal/web/templates/*")
	
	auth.Setup(r)

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "login.html", nil)
	})
	r.GET("/auth/google", web.AuthLogin)
	r.GET("/auth/google/callback", web.AuthCallback)

	authorized := r.Group("/")
	authorized.Use(auth.AuthMiddleware())
	authorized.Use(csrf.Middleware(csrf.Options{
		Secret: os.Getenv("SESSION_SECRET"),
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))
	{
		authorized.GET("/dashboard", web.Dashboard)
		authorized.POST("/upload", web.Upload)
	}

	r.Static("/sites", "./data/sites")

	log.Println("Server running on :8080")
	r.Run(":8080")
}
