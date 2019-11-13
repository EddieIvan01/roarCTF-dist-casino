package route

import (
	"casino/app"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var (
	Done = false
)

type Status struct {
	Players  []string `json:"players"`
	Pendings []string `json:"pendings"`
	Winners  []string `json:"winners"`
}

// returns status: players, pendings, winners
func PlayerStatus(c *gin.Context) {
	players := make([]string, 0)
	pendings := make([]string, 0)
	winners := make([]string, 0)

	for u, _ := range app.CasinoService.Players {
		players = append(players, u)
	}
	for u, _ := range app.CasinoService.Pendings {
		pendings = append(pendings, u)
	}
	for u, _ := range app.CasinoService.Winners {
		winners = append(winners, u)
	}

	c.JSON(200, Msg{0, "ok", Status{
		players,
		pendings,
		winners,
	}})
}

func ServiceStatus(c *gin.Context) {
	status := app.CasinoService.Doing
	c.JSON(200, Msg{0, "ok", gin.H{
		"status": status,
	}})
}

/******************************************************
*  admin required
*******************************************************/
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		uname := session.Get("uname").(string)
		isAdmin := session.Get("isAdmin").(bool)

		if uname == "admin" && isAdmin {
			c.Next()
		} else {
			c.AbortWithStatus(403)
		}
	}
}

// change user's state from pending to formal player
func AddPlayer(c *gin.Context) {
	u := struct {
		Uname string `json:"uname"`
	}{}
	err := c.ShouldBind(&u)

	if err != nil {
		c.JSON(200, Msg{-1, "error", nil})
		return
	}

	if user, ok := app.Users[u.Uname]; ok {
		err = app.CasinoService.AddPlayer(user)
		if err != nil {
			c.JSON(200, Msg{-1, err.Error(), nil})
			return
		}
	}
	c.JSON(200, Msg{0, "ok", nil})
}

func ServiceStart(c *gin.Context) {
	app.CasinoService.Start()
	c.JSON(200, Msg{0, "ok", nil})
}

func ServiceReset(c *gin.Context) {
	app.CasinoService.Reset()
	c.JSON(200, Msg{0, "ok", nil})
}
