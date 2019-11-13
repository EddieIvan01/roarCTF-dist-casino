package route

import (
	"casino/app"
	"crypto/sha1"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	padding               = "0123456789qwertyuioplkjhgfdsazxcvbnm<>?!@#$%^&*()_+"
	MAX_HASH_TABLE_LENGTH = 200
)

var (
	globalUserHash = make(map[string]string)

	ErrTokenNotExists = errors.New("token dont exist")
	ErrHashDontMatch  = errors.New("hash dont match")
)

func getRandomStr(n int) []byte {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = padding[rand.Intn(51)]
	}
	return b
}

func setHashTable(k string, v string) {
	if len(globalUserHash) >= MAX_HASH_TABLE_LENGTH {
		for key, _ := range globalUserHash {
			delete(globalUserHash, key)
			break
		}
	}
	globalUserHash[k] = v
}

func sha1Hash(b []byte) string {
	s := sha1.New()
	s.Write(b)
	return fmt.Sprintf("%x", s.Sum(nil))
}

func getHashAndToken() (string, string) {
	x := getRandomStr(8)
	hash := sha1Hash(x)
	token := sha1Hash(getRandomStr(8))
	setHashTable(token, hash[:6])

	// for debug
	fmt.Println(string(x))
	return hash[:6], token
}

func HashHandler(c *gin.Context) {
	hash, token := getHashAndToken()
	c.SetCookie("hash-token", token, 3600, "/", "", false, false)
	c.JSON(200, Msg{
		0,
		"ok",
		hash,
	})
}

type RegisterJSON struct {
	LoginJSON
	RawHash string `json:"raw_hash"`
}

type LoginJSON struct {
	Uname string `json:"uname"`
	Pwd   string `json:"pwd"`
}

func RegisterHandler(c *gin.Context) {
	j := RegisterJSON{}
	err := c.ShouldBindJSON(&j)
	if err != nil {
		c.JSON(200, Msg{-1, "param error", nil})
		return
	}

	token, err := c.Cookie("hash-token")
	if err != nil {
		c.JSON(200, Msg{-1, "hash token error", nil})
		return
	}
	c.SetCookie("hash-token", "x", -1, "/", "", false, false)

	hash, found := globalUserHash[token]
	if !found {
		c.JSON(200, Msg{-1, ErrTokenNotExists.Error(), nil})
		return
	}
	delete(globalUserHash, token)

	if sha1Hash([]byte(j.RawHash))[:6] != hash {
		c.JSON(200, Msg{-1, ErrHashDontMatch.Error(), nil})
		return
	}

	pwdHash := sha1Hash([]byte(j.Pwd + SALT))
	s, err := DB.Prepare("INSERT INTO users(uname, pwd) VALUES(?, ?);")
	s.Exec(j.Uname, pwdHash)
	if err != nil {
		c.JSON(200, Msg{-1, "register error", nil})
		return
	}

	c.JSON(200, Msg{0, "ok", nil})
}

func LoginHandler(c *gin.Context) {
	j := LoginJSON{}
	err := c.ShouldBindJSON(&j)
	if err != nil {
		c.JSON(200, Msg{-1, "param error", nil})
		return
	}

	// double insurance
	// also protected by WAF
	for _, w := range []string{"delete", "insert", "update"} {
		if strings.Contains(strings.ToLower(j.Uname), w) {
			c.JSON(200, Msg{-1, "something wrong", nil})
			return
		}
	}

	// may be SQLi here
	// but dont worry, baby hackers cant break my waf
	rows, err := DB.Query(fmt.Sprintf("SELECT pwd FROM users WHERE uname='%s';", j.Uname))
	if err != nil {
		c.JSON(200, Msg{-1, err.Error(), nil})
		return
	}
	defer rows.Close()

	var pwd string
	if rows.Next() {
		rows.Scan(&pwd)
	}
    fmt.Println(pwd)
	if sha1Hash([]byte(j.Pwd+SALT)) != pwd {
		c.JSON(200, Msg{-1, "passwd wrong", nil})
		return
	}

	session := sessions.Default(c)
	session.Set("uname", j.Uname)
	session.Set("isAdmin", false)
	session.Save()

	// global user list
	UserAuth[j.Uname] = j.Pwd
	u := app.NewUser(j.Uname)
	app.Users[j.Uname] = u

	c.JSON(200, Msg{0, "ok", nil})
}
