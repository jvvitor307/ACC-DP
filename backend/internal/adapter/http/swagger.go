package http

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed swagger/openapi.yaml
var swaggerFS embed.FS

var swaggerHTMLTemplate = template.Must(template.New("swagger").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>ACC-DP Backend Swagger</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
    <style>
      body {
        margin: 0;
        background: #fafafa;
      }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.ui = SwaggerUIBundle({
        url: "{{ .SpecURL }}",
        dom_id: "#swagger-ui",
        deepLinking: true,
      });
    </script>
  </body>
</html>`))

type swaggerPageData struct {
	SpecURL string
}

func registerSwaggerRoutes(router *gin.Engine) {
	router.GET("/swagger", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}

		specURL := fmt.Sprintf("%s://%s/swagger/openapi.yaml", scheme, c.Request.Host)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := swaggerHTMLTemplate.Execute(c.Writer, swaggerPageData{SpecURL: specURL}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render swagger ui"})
		}
	})

	router.GET("/swagger/openapi.yaml", func(c *gin.Context) {
		content, err := swaggerFS.ReadFile("swagger/openapi.yaml")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load openapi spec"})
			return
		}

		c.Data(http.StatusOK, "application/yaml; charset=utf-8", content)
	})
}
