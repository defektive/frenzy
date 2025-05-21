package runner

import (
	"bytes"
	"fmt"
	"github.com/defektive/frenzy/pkg/common"
	"github.com/defektive/frenzy/pkg/server"
	"text/template"
)

func Run() {

}

type Rule struct {
	Name    string `yaml:"Name"`
	Module  string `yaml:"Module"`
	Search  string `yaml:"Search"`
	Replace string `yaml:"Replace"`

	template *template.Template
}

func (r Rule) Precompile() {
	testTemplate, err := template.New(fmt.Sprintf("%s/%s", r.Name, r.Replace)).Funcs(template.FuncMap{
		"hasPermission": func(feature string) bool {
			return false
		},
	}).Parse(r.Search)
	common.EnsureNotError(err)

	r.template = testTemplate
}

func (r Rule) Send(in *bytes.Buffer) (err error) {
	return r.template.Execute(in, server.GetConfig().Serve)

}

func (r Rule) Recv(in *bytes.Buffer) (err error) {
	return r.template.Execute(in, server.GetConfig().Serve)
}
