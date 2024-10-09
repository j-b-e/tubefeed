package app

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /tab/:id
func (a App) edittab(c *gin.Context) {
	ctx := c.Request.Context()
	tabid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	tabs, err := a.Db.LoadTabs(ctx)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.HTML(http.StatusOK, "tabedit.html", gin.H{"Tab": tabid, "Name": tabs[tabid]})
}

// PATCH /tab/:id
func (a App) patchtab(c *gin.Context) {
	ctx := c.Request.Context()
	tabid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	newname := c.PostForm("name")
	log.Printf("newname: %s", newname)

	err = a.Db.ChangeTabName(ctx, tabid, newname)
	if err != nil {
		// ignore err, old name will be reused
		log.Println(err)
	}
	tabs, err := a.Db.LoadTabs(ctx)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.HTML(http.StatusOK, "tablist.html", gin.H{"Tabs": tabs, "tab": tabid})
}

// GET /tab and GET /tab/:id
func (a App) tablist(c *gin.Context) {
	ctx := c.Request.Context()
	ret := c.Param("id")
	if ret == "" {
		ret = "1"
	}
	active, err := strconv.Atoi(ret)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	tabs, err := a.Db.LoadTabs(ctx)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.HTML(http.StatusOK, "tablist.html", gin.H{"Tabs": tabs, "tab": active})
}

// POST /tab -- create a new tab
func (a App) createtab(c *gin.Context) {
	ctx := c.Request.Context()
	if err := a.Db.AddTab(ctx, "New Tab"); err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	a.tablist(c)
}

// DELETE /tab/:id
func (a App) deleteTab(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	// TODO: Cleanup Audio Files
	err = a.Db.DeleteTab(ctx, id)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	a.tablist(c)
}
