package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"opennamu/route"
)

func View_clan_routes(r *gin.Engine) {
	r.GET("/clan/sidebar", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_sidebar(Make_route_config(c))))
	})
	r.GET("/clan/hall", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		Write_data(c, http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_hall(Make_route_config(c), nil)))
	})
	r.POST("/clan/hall", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
		if c.Request.ParseForm() != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.Header("Cache-Control", "no-store")
		Write_data(c, http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_hall(Make_route_config(c), c.Request.PostForm)))
	})
	r.GET("/new-document", func(c *gin.Context) {
		Write_data(c, http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_document(Make_route_config(c), nil)))
	})
	r.POST("/new-document", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
		if c.Request.ParseForm() != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		Write_data(c, http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_document(Make_route_config(c), c.Request.PostForm)))
	})
	r.GET("/clan/users", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		Write_data(c, http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_users(Make_route_config(c), nil)))
	})
	r.POST("/clan/users", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16384)
		if c.Request.ParseForm() != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.Header("Cache-Control", "no-store")
		Write_data(c, http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_users(Make_route_config(c), c.Request.PostForm)))
	})
}
