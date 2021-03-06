package elastic

import (
	"bytes"
	"io"
	"text/template"

	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

// Helper functions used within the package

// format parses text as as a template and excute it with the data
func format(text string, data interface{}) (*bytes.Buffer, error) {
	op := "format"
	var buf bytes.Buffer
	if err := template.Must(template.New("").Parse(text)).Execute(&buf, data); err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	return &buf, nil
}

// read reads the data from r and returns it as a string
func read(r io.Reader) string {
	var b bytes.Buffer
	b.ReadFrom(r)
	return b.String()
}
