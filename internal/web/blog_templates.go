package web

import (
	"html/template"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamanishx/xserve/internal/db"
	"github.com/iamanishx/xserve/internal/engine"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, extension.Linkify, extension.TaskList),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

func init() {
	RegisterTemplateFuncs()
}

var templateFuncs = template.FuncMap{
	"now":          Now,
	"calcReadTime": CalcReadTime,
	"truncate":     Truncate,
	"renderMarkdown": RenderMarkdown,
}

func RegisterTemplateFuncs() {
	for name, fn := range templateFuncs {
		ginFuncs[name] = fn
	}
}

var ginFuncs = template.FuncMap{}

func AddTemplateFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		ginFuncs[k] = v
	}
}

func CalcReadTime(content string) int {
	words := strings.Fields(content)
	avgWordsPerMin := 200
	minutes := len(words) / avgWordsPerMin
	if minutes < 1 {
		return 1
	}
	return minutes
}

func Truncate(content string, maxLen int) string {
	plain := stripMarkdown(content)
	if len(plain) <= maxLen {
		return plain
	}
	return strings.TrimSpace(plain[:maxLen]) + "..."
}

func Now() time.Time {
	return time.Now()
}

func RenderMarkdown(content string) template.HTML {
	var buf strings.Builder
	if err := md.Convert([]byte(content), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(content))
	}
	return template.HTML(buf.String())
}

func stripMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"#", "", "##", "", "###", "", "####", "", "#####", "", "######", "",
		"**", "", "*", "", "__", "", "_", "",
		"`", "", "```", "",
		"[", "", "](", "", ")", "",
		"![", "", "](", "", ")", "",
		"- ", "", "* ", "", "+ ", "",
		">", "", "---", "", "___", "",
		"\n", " ", "\r", "",
	)
	return replacer.Replace(s)
}

type BlogIndexData struct {
	User   *db.User
	Posts  []*db.Post
}

type BlogPostData struct {
	User    *db.User
	Post    *db.Post
	Content template.HTML
	FullURL string
}

func BlogIndex(c *gin.Context) {
	userID := c.Param("userID")

	user, err := db.GetUser(userID)
	if err != nil {
		c.String(404, "User not found")
		return
	}

	posts, err := db.GetPublicPosts(user.ID)
	if err != nil {
		c.String(500, "Internal server error")
		return
	}

	data := BlogIndexData{
		User:  user,
		Posts: posts,
	}

	c.HTML(200, "blog_index.html", data)
}

func BlogPost(c *gin.Context) {
	userID := c.Param("userID")
	slug := c.Param("slug")

	user, err := db.GetUser(userID)
	if err != nil {
		c.String(404, "User not found")
		return
	}

	post, err := db.GetPublicPostBySlug(user.ID, slug)
	if err != nil {
		c.String(404, "Post not found")
		return
	}

	content := RenderMarkdown(post.Content)

	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := c.Request.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	fullURL := scheme + "://" + host + "/" + user.ID + "/" + post.Slug

	data := BlogPostData{
		User:    user,
		Post:    post,
		Content: content,
		FullURL: fullURL,
	}

	c.HTML(200, "blog_post.html", data)
}

func BlogPostPreview(c *gin.Context) {
	uid := c.GetString("user_id")
	slug := c.Param("slug")

	post, err := db.GetPostBySlug(uid, slug)
	if err != nil {
		c.String(404, "Post not found")
		return
	}

	user, err := db.GetUser(uid)
	if err != nil {
		c.String(404, "User not found")
		return
	}
	content := RenderMarkdown(post.Content)

	data := BlogPostData{
		User:    user,
		Post:    post,
		Content: content,
		FullURL: "",
	}

	c.HTML(200, "blog_post.html", data)
}

func BuildPost(userID string, post *db.Post) error {
	content := RenderMarkdown(post.Content)
	return engine.BuildPublicPost(userID, post.Slug, post.Title, string(content))
}
