package engine

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"path/filepath"
	"sync"

	"github.com/iamanishx/xserve/internal/storage"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

const defaultTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{ .Title }}</title>
<style>body{font-family:sans-serif;max-width:800px;margin:0 auto;padding:20px;line-height:1.6}img{max-width:100%}</style>
</head>
<body>
{{ .Content }}
</body>
</html>
`

var r2Client *storage.R2Client

func InitStorage() {
	client, err := storage.NewR2Client()
	if err != nil {
		r2Client = nil
		return
	}
	r2Client = client
}

func BuildSite(userID string, files map[string][]byte) error {
	var outDir string
	useR2 := r2Client != nil

	if useR2 {
		outDir = "sites/" + userID
	} else {
		outDir = filepath.Join("data", "sites", userID)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return err
		}
	}

	tmpl, err := template.New("base").Parse(defaultTemplate)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(files))
	var fileList []string

	for name, content := range files {
		wg.Add(1)
		go func(name string, content []byte) {
			defer wg.Done()
			var buf bytes.Buffer
			if err := md.Convert(content, &buf); err != nil {
				errChan <- err
				return
			}

			htmlContent := buf.String()
			baseName := name
			if filepath.Ext(name) == ".md" {
				baseName = name[:len(name)-3]
			}
			outName := baseName + ".html"
			key := outDir + "/" + outName

			data := struct {
				Title   string
				Content template.HTML
			}{
				Title:   baseName,
				Content: template.HTML(htmlContent),
			}

			var buf2 bytes.Buffer
			if err := tmpl.Execute(&buf2, data); err != nil {
				errChan <- err
				return
			}

			if useR2 {
				if err := r2Client.Upload(context.Background(), key, buf2.Bytes(), "text/html"); err != nil {
					errChan <- err
					return
				}
			} else {
				f, err := os.Create(filepath.Join(outDir, outName))
				if err != nil {
					errChan <- err
					return
				}
				buf2.WriteTo(f)
				f.Close()
			}
		}(name, content)

		baseName := name
		if filepath.Ext(name) == ".md" {
			baseName = name[:len(name)-3]
		}
		fileList = append(fileList, baseName+".html")
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	indexHTML := `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Your Site</title><style>body{font-family:sans-serif;max-width:800px;margin:2rem auto;padding:20px}h1{color:#333}ul{list-style:none;padding:0}li{margin:10px 0}a{color:#6c5ce7;text-decoration:none;font-size:18px}a:hover{text-decoration:underline}</style></head><body><h1>Your Pages</h1><ul>`
	for _, f := range fileList {
		indexHTML += `<li><a href="` + f + `">` + f + `</a></li>`
	}
	indexHTML += `</ul></body></html>`

	indexKey := outDir + "/index.html"

	if useR2 {
		return r2Client.Upload(context.Background(), indexKey, []byte(indexHTML), "text/html")
	}

	return os.WriteFile(filepath.Join(outDir, "index.html"), []byte(indexHTML), 0644)
}

func BuildPublicPost(userID, slug, title, htmlContent string) error {
	if r2Client == nil {
		return buildPostLocalHTML(userID, slug, title, htmlContent)
	}
	return buildPostR2HTML(userID, slug, title, htmlContent)
}

func buildPostR2HTML(userID, slug, title, htmlContent string) error {
	fullTemplate := `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{ .Title }}</title>
<style>
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:720px;margin:0 auto;padding:2rem;line-height:1.8;color:#1a1a1a}
h1,h2,h3{color:#111;margin-top:2rem}
a{color:#0066cc}
img{max-width:100%;border-radius:8px}
code{background:#f5f5f5;padding:0.2rem 0.4rem;border-radius:4px;font-size:0.9em}
pre{background:#1e1e1e;color:#d4d4d4;padding:1rem;border-radius:8px;overflow-x:auto}
pre code{background:none;padding:0}
blockquote{border-left:4px solid #ddd;margin:1rem 0;padding-left:1rem;color:#666}
</style>
</head>
<body>
<h1>{{ .Title }}</h1>
{{ .Content }}
</body>
</html>
`

	tmpl, err := template.New("post").Parse(fullTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	data := struct {
		Title   string
		Content template.HTML
	}{
		Title:   title,
		Content: template.HTML(htmlContent),
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	key := "sites/" + userID + "/" + slug + ".html"
	return r2Client.Upload(context.Background(), key, buf.Bytes(), "text/html")
}

func buildPostLocalHTML(userID, slug, title, htmlContent string) error {
	outDir := filepath.Join("data", "sites", userID)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	fullTemplate := `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{ .Title }}</title>
<style>
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:720px;margin:0 auto;padding:2rem;line-height:1.8;color:#1a1a1a}
h1,h2,h3{color:#111;margin-top:2rem}
a{color:#0066cc}
img{max-width:100%;border-radius:8px}
code{background:#f5f5f5;padding:0.2rem 0.4rem;border-radius:4px;font-size:0.9em}
pre{background:#1e1e1e;color:#d4d4d4;padding:1rem;border-radius:8px;overflow-x:auto}
pre code{background:none;padding:0}
blockquote{border-left:4px solid #ddd;margin:1rem 0;padding-left:1rem;color:#666}
</style>
</head>
<body>
<h1>{{ .Title }}</h1>
{{ .Content }}
</body>
</html>
`

	tmpl, err := template.New("post").Parse(fullTemplate)
	if err != nil {
		return err
	}

	outFile := filepath.Join(outDir, slug+".html")
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		Title   string
		Content template.HTML
	}{
		Title:   title,
		Content: template.HTML(htmlContent),
	}

	return tmpl.Execute(f, data)
}
