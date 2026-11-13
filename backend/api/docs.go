// Package apidocs embeds the OpenAPI description of the API and serves it
// with Swagger UI. The description is checked against the registered routes
// by package speccheck, so it cannot drift from the handlers.
package apidocs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Spec is the OpenAPI document, api/swagger.yaml.
//
//go:embed swagger.yaml
var Spec []byte

// Paths served by Register. They are not part of the described API.
const (
	UIPath   = "/api/docs"
	SpecPath = "/api/docs/openapi.yaml"
)

const uiPage = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Approval Engine API</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui.css" integrity="sha384-wxLW6kwyHktdDGr6Pv1zgm/VGJh99lfUbzSn6HNHBENZlCN7W602k9VkGdxuFvPn" crossorigin="anonymous">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui-bundle.js" integrity="sha384-wmyclcVGX/WhUkdkATwhaK1X1JtiNrr2EoYJ+diV3vj4v6OC5yCeSu+yW13SYJep" crossorigin="anonymous"></script>
<script>
window.ui = SwaggerUIBundle({ url: "/api/docs/openapi.yaml", dom_id: "#swagger-ui", deepLinking: true });
</script>
</body>
</html>
`

// Register serves the Swagger UI page and the raw document. Both are public.
func Register(router gin.IRoutes) {
	router.GET(UIPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(uiPage))
	})
	router.GET(SpecPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", Spec)
	})
}
