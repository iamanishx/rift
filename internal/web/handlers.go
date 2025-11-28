package web

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamanishx/xserve/internal/auth"
	"github.com/iamanishx/xserve/internal/db"
	"github.com/iamanishx/xserve/internal/engine"
	"github.com/markbates/goth/gothic"
)

func AuthCallback(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", "google")
	c.Request.URL.RawQuery = q.Encode()
	
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.String(400, "Auth failed: "+err.Error())
		return
	}

	u := &db.User{
		ID:        user.UserID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		CreatedAt: time.Now(),
	}
	if err := db.SaveUser(u); err != nil {
		c.String(500, "Database error: "+err.Error())
		return
	}

	token, err := auth.GenerateJWT(u.ID)
	if err != nil {
		c.String(500, "Token generation failed")
		return
	}

	c.SetCookie("auth_token", token, 30*24*60*60, "/", "", false, true)
	c.Redirect(302, "/dashboard")
}

func AuthLogin(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", "google")
	c.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func Dashboard(c *gin.Context) {
	uid := c.GetString("user_id")
	user, _ := db.GetUser(uid)
	c.HTML(200, "dashboard.html", gin.H{
		"User": user,
	})
}

func Upload(c *gin.Context) {
	form, _ := c.MultipartForm()
	files := form.File["files"]
	
	fileMap := make(map[string][]byte)
	for _, file := range files {
		f, _ := file.Open()
		defer f.Close()
		content, _ := io.ReadAll(f)
		fileMap[file.Filename] = content
	}

	uid := c.GetString("user_id")

	if err := engine.BuildSite(uid, fileMap); err != nil {
		c.String(500, "Build failed")
		return
	}

	c.Redirect(302, "/dashboard")
}
