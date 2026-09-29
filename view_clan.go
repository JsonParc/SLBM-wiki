package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"opennamu/route"
)

const clan_backup_max_upload = 512 * 1024 * 1024

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
	r.GET("/clan/backup", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		Write_data(c, http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_backup(Make_route_config(c), false, "", nil, "")))
	})
	r.GET("/clan/backup/download", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		raw, name, err := route.Api_clan_backup_export(Make_route_config(c))
		if err != nil {
			message := "clan_backup_failed"
			if err.Error() == "auth" || err.Error() == "clan_backup_sqlite_only" {
				message = err.Error()
			}
			Write_data(c, http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_backup(Make_route_config(c), false, "", nil, message)))
			return
		}
		c.Header("Content-Disposition", `attachment; filename="`+name+`"`)
		c.Data(http.StatusOK, "application/zip", raw)
	})
	r.POST("/clan/backup", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, clan_backup_max_upload)
		file, _, err := c.Request.FormFile("backup")
		if err != nil {
			Write_data(c, http.StatusBadRequest, "text/html; charset=utf-8", []byte(route.View_clan_backup(Make_route_config(c), false, "", nil, "clan_backup_invalid")))
			return
		}
		defer file.Close()
		raw, err := io.ReadAll(file)
		if err != nil {
			Write_data(c, http.StatusBadRequest, "text/html; charset=utf-8", []byte(route.View_clan_backup(Make_route_config(c), false, "", nil, "clan_backup_invalid")))
			return
		}
		Write_data(c, http.StatusOK, "text/html; charset=utf-8", []byte(route.View_clan_backup(Make_route_config(c), true, c.Request.PostFormValue("csrf"), raw, "")))
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
