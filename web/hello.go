package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/javitab/go-web/templates"
	"github.com/noirbizarre/gonja"
)

// Parse the template

func helloHandler(c *gin.Context) {
	// Data to pass to the template
	message := c.Request.URL.Query().Get("message")
	tmpl, _ := templates.GetTemplate("base.j2")

	// Execute the template with the message
	out, err := tmpl.Execute(gonja.Context{"message": message})

	// Check for errors
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Return the rendered template
	c.Data(http.StatusOK, "text/html", []byte(out))
}
