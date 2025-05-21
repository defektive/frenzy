package modules

import "text/template"

type Module interface {
	Name() string
}

var TemplateFuncMap = template.FuncMap{
	"hasPermission": func(feature string) bool {
		return false
	},
}
