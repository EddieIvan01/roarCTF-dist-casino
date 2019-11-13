package route

import (
	"casino/app"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

/******************************************************
*  normal user
*******************************************************/
func UserInfo(c *gin.Context) {
	session := sessions.Default(c)
	uname := session.Get("uname").(string)

	var flag string
	u, ok := app.Users[uname]
	u.TotalBalance = u.CalcBalance()
	if !ok {
		c.JSON(200, Msg{-1, "error", nil})
		return
	}

	_, win := app.CasinoService.Winners[uname]
	if u.TotalBalance > 999999 || win {
		flag = os.Getenv("flag")
	} else {
		flag = "nothing to show, to be a winner"
	}
	c.JSON(200, Msg{0, flag, u})
}

// applicat to be a player
// add self to pending list
func ApplicateToCasino(c *gin.Context) {
	session := sessions.Default(c)
	uname := session.Get("uname").(string)

	if u, ok := app.Users[uname]; ok {
		app.CasinoService.ApplicatUser(u)
		c.JSON(200, Msg{0, "ok", nil})
		return
	}
	c.JSON(200, Msg{-1, "something wrong", nil})
}

// beg for random amount money
// could only beg 6 times or less
func UserBeg(c *gin.Context) {
	session := sessions.Default(c)
	uname := session.Get("uname").(string)

	if u, ok := app.Users[uname]; ok {
		u.Beg()
		app.Users[uname] = u
		c.JSON(200, Msg{0, "ok", gin.H{
			"ok": ok,
		}})
		return
	}
	c.JSON(200, Msg{-1, "something wrong", nil})
}

// reset your account
func UserReset(c *gin.Context) {
	session := sessions.Default(c)
	uname := session.Get("uname").(string)

	if u, ok := app.Users[uname]; ok {
		u.Reset()
		app.Users[uname] = u
		c.JSON(200, Msg{0, "ok", nil})
		return
	}
	c.JSON(200, Msg{-1, "something wrong", nil})
}
