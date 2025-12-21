package utils

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
)

// RenderTemplate renders a Pongo2 template with the global context
func RenderTemplate(c *gin.Context, templatePath string, extraContext map[string]interface{}) {
	ctx := GetGlobalContext(c)

	// Merge extra context
	if extraContext != nil {
		for k, v := range extraContext {
			ctx[k] = v
		}
	}

	tpl, err := pongo2.FromFile(templatePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template Load Error: "+err.Error())
		return
	}

	out, err := tpl.Execute(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template Execute Error: "+err.Error())
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(out))
}
