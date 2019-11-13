package main

import (
	"casino/route"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/kylelemons/go-gypsy/yaml"
)

var (
	domain      string
	authPort    string
	servicePort string
	salt        string
	secretKey   string
	adminPwd    string
)

func init() {
	config, err := yaml.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("config file dont exists")
	}

	domain, err = config.Get("DOMAIN")
	if err != nil {
		log.Fatal(err.Error())
	}

	authPort, err = config.Get("AUTH_PORT")
	if err != nil {
		log.Fatal(err.Error())
	}
	servicePort, err = config.Get("SERVICE_PORT")
	if err != nil {
		log.Fatal(err.Error())
	}

	/*
		authPort = os.Getenv("auth")
		servicePort = os.Getenv("service")
	*/

	salt, err = config.Get("SALT")
	if err != nil {
		log.Fatal(err.Error())
	}
	secretKey, err = config.Get("SECRET_KEY")
	if err != nil {
		log.Fatal(err.Error())
	}
	adminPwd, err = config.Get("ADMIN_PWD")
	if err != nil {
		log.Fatal(err.Error())
	}

	err = route.InitDB(salt, secretKey, adminPwd)
	if err != nil {
		log.Println(err.Error())
	}

}

func main() {
	defer route.DB.Close()
	gin.SetMode(gin.ReleaseMode)
	authRouter := gin.Default()
	serviceRouter := gin.Default()

	authRouter.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(string) bool { return true },
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	serviceRouter.Use(cors.New(cors.Config{
		AllowOriginFunc:  func(string) bool { return true },
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	authRouter.Use(secure.New(secure.Config{
		STSSeconds:            315360000,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
		IENoOpen:              true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))
	serviceRouter.Use(secure.New(secure.Config{
		STSSeconds:            315360000,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
		IENoOpen:              true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	storage := cookie.NewStore([]byte(secretKey))
	authRouter.Use(sessions.Sessions("casino", storage))
	serviceRouter.Use(sessions.Sessions("casino", storage))

	// auth router will be protected by WAF
	auth := authRouter.Group("/auth")
	auth.GET("/hash", route.HashHandler)
	auth.POST("/register", route.RegisterHandler)
	auth.POST("/login", route.LoginHandler)

	api := serviceRouter.Group("/api")

	user := api.Group("/u")
	user.Use(route.SensitiveAuth())
	user.POST("/info", route.UserInfo)
	user.POST("/beg", route.UserBeg)
	user.POST("/reset", route.UserReset)
	user.POST("/join", route.ApplicateToCasino)

	service := api.Group("/service")
	service.GET("/player-status", route.PlayerStatus)
	service.GET("/status", route.ServiceStatus)
	serviceManage := service.Group("/manage")
	serviceManage.Use(route.AdminRequired())
	serviceManage.GET("/start", route.ServiceStart)
	serviceManage.GET("/reset", route.ServiceReset)
	serviceManage.POST("/add-player", route.AddPlayer)

	go authRouter.Run("0.0.0.0:" + authPort)
	serviceRouter.Run("0.0.0.0:" + servicePort)
}
