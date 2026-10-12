package response

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(
		200,
		gin.H{
			"success": true,
			"message": message,
			"data":    data,
			"error":   nil,
		},
	)
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(
		201,
		gin.H{
			"success": true,
			"message": message,
			"data":    data,
			"error":   nil,
		},
	)
}

// Error writes an error response. Server errors (5xx) are logged with the
// request id and replaced by a generic message, so database or internal
// details never reach the client.
func Error(c *gin.Context, status int, message string, err interface{}) {
	if status >= http.StatusInternalServerError {
		log.Printf("request_id=%s path=%s status=%d error=%q detail=%v",
			c.GetString("request_id"), c.FullPath(), status, message, err)
		message = "something went wrong, please try again"
		err = nil
	}
	c.JSON(
		status,
		gin.H{
			"success": false,
			"message": message,
			"data":    nil,
			"error":   err,
		},
	)
}
