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
			outName := name + ".html"
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
				Title:   name,
				Content: template.HTML(htmlContent),
			}

			if err := tmpl.Execute(f, data); err != nil {
				errChan <- err
				return
			}
		}(name, content)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}
	return nil
}
