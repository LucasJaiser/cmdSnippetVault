package templatevar

import (
	"bytes"
	"regexp"
	"text/template"
)

func Parse(command string) (*template.Template, []string, error) {

	tmpl, err := template.New("cmd").Parse(command)
	var keys []string

	if err != nil {
		return nil, nil, err
	}

	re := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)
	matches := re.FindAllStringSubmatch(command, -1)
	for _, m := range matches {
		keys = append(keys, m[1])
	}

	return tmpl, keys, nil
}

func Resolve(tmpl *template.Template, values *map[string]string) (string, error) {

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, values)

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
