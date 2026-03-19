package web

import (
	"github.com/gin-gonic/gin"
	"github.com/iamanishx/xserve/internal/db"
)

type EditorData struct {
	Draft *db.Draft
	User  *db.User
}

func Editor(c *gin.Context) {
	uid := c.GetString("user_id")
	user, _ := db.GetUser(uid)

	var data EditorData
	data.User = user

	id := c.Param("id")
	if id != "" {
		if d, err := db.GetDraft(id); err == nil && d.UserID == uid {
			data.Draft = d
		}
	}

	c.HTML(200, "editor.html", data)
}
