package templates

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateEmbeds(t *testing.T) {
	templateLoaded := false
	_, err := GetTemplate("hello.html")
	if err == nil {
		fmt.Println("Template loaded successfully")
		templateLoaded = true
	}

	assert.Equal(t, templateLoaded, true)

}
