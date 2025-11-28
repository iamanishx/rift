package web

import (
	"io"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/iamanishx/xserve/internal/db"
	"github.com/iamanishx/xserve/internal/engine"
	"github.com/markbates/goth/gothic"
)

func AuthCallback(c *gin.Context) {
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.String(400, "Auth failed")
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
		c.String(500, "Database error")
		return
	}

	s := sessions.Default(c)
	s.Set("user_id", u.ID)
	s.Save()

	c.Redirect(302, "/dashboard")
}

func AuthLogin(c *gin.Context) {
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func Dashboard(c *gin.Context) {
	s := sessions.Default(c)
	uid := s.Get("user_id").(string)
	user, _ := db.GetUser(uid)
	c.HTML(200, "dashboard.html", gin.H{
		"User": user,
		"CSRF": c.GetString("csrf_token"),
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

	s := sessions.Default(c)
	uid := s.Get("user_id").(string)

	if err := engine.BuildSite(uid, fileMap); err != nil {
		c.String(500, "Build failed")
		return
	}

	c.Redirect(302, "/dashboard")
}
