package engine

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"sync"

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

func BuildSite(userID string, files map[string][]byte) error {
	outDir := filepath.Join("data", "sites", userID)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
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
			outFile := filepath.Join(outDir, outName)

			f, err := os.Create(outFile)
			if err != nil {
				errChan <- err
				return
			}
			defer f.Close()

			data := struct {
				Title   string
				Content template.HTML
			}{
				Title:   baseName,
				Content: template.HTML(htmlContent),
			}

			if err := tmpl.Execute(f, data); err != nil {
				errChan <- err
				return
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

	entries, err := os.ReadDir(outDir)
	if err == nil {
		fileList = []string{}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".html" && entry.Name() != "index.html" {
				fileList = append(fileList, entry.Name())
			}
		}
	}

	indexHTML := `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Your Site</title><style>body{font-family:sans-serif;max-width:800px;margin:2rem auto;padding:20px}h1{color:#333}ul{list-style:none;padding:0}li{margin:10px 0}a{color:#6c5ce7;text-decoration:none;font-size:18px}a:hover{text-decoration:underline}</style></head><body><h1>Your Pages</h1><ul>`
	for _, f := range fileList {
		indexHTML += `<li><a href="` + f + `">` + f + `</a></li>`
	}
	indexHTML += `</ul></body></html>`
	
	return os.WriteFile(filepath.Join(outDir, "index.html"), []byte(indexHTML), 0644)
}
