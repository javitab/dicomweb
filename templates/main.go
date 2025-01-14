package templates

import (
	"embed"
	"fmt"

	"github.com/noirbizarre/gonja"
	"github.com/noirbizarre/gonja/exec"
)

//go:embed *
var templates embed.FS

func GetTemplate(name string) (*exec.Template, error) {
	template, err := templates.ReadFile(name)
	if err != nil {
		fmt.Printf("Unable to load template %v", name)
		return nil, err
	} //
	tmpl := gonja.Must(gonja.FromBytes(template))
	return tmpl, err
}
