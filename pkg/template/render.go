package template

import (
	"bytes"
	"text/template"
)

func Render(name, content string, data interface{}, funcs template.FuncMap) (string, error) {
	tpl, err := template.New(name).
		Funcs(funcs).
		Parse(content)
	if err != nil {
		return "", err
	}
	wr := &bytes.Buffer{}
	err = tpl.Option("missingkey=error").Execute(wr, data)
	if err != nil {
		return "", err
	}
	return wr.String(), nil
}
