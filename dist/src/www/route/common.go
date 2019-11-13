package route

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// serialize data buffer
type Msg struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

type User struct {
	Uname string
	Pwd   string
}

type Auth struct {
	Pwd string `json:"pwd"`
}

// for sensitive auth
var UserAuth = make(map[string]string)

func SensitiveAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		a := Auth{}
		err := c.ShouldBindJSON(&a)
		if err != nil {
			c.JSON(200, Msg{-1, "please auth before sensitive operating", nil})
			c.AbortWithStatus(401)
		}

		session := sessions.Default(c)
		uname, ok := session.Get("uname").(string)
		if !ok {
			c.JSON(200, Msg{-1, "login first", nil})
			c.AbortWithStatus(401)
		}

		if a.Pwd != UserAuth[uname] {
			c.JSON(200, Msg{-1, "auth fail", nil})
			c.AbortWithStatus(401)
		}

		c.Next()
	}
}
