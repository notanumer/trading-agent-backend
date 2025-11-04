// @title           DeepSeek Trader API
// @version         1.0
// @description     This is a DeepSeek Trader API.
// @termsOfService  http://swagger.io/terms/

// @contact.name   DeepSeek Trader
// @contact.email  your@email.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

package api

import (
	"net/http"

	"deepseek-trader/api/handlers"
	"deepseek-trader/api/middleware"
	"deepseek-trader/config"

	"github.com/gin-gonic/gin"

	docs "deepseek-trader/api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	basePath = "/api/v1"
)

func NewRouter(handlers *handlers.Handler, cfg *config.Settings) http.Handler {
	r := gin.New()
	docs.SwaggerInfo.BasePath = basePath

	r.Use(gin.Recovery())
	r.Use(middleware.Cors())

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// health check
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := r.Group(basePath)

	// Auth
	api.POST("/auth/register", handlers.Register)
	api.POST("/auth/login", handlers.Login)

	secured := api.Group("")
	secured.Use(middleware.Aut(cfg.JWTSecret))

	secured.GET("/me", handlers.Me)

	// Wallet
	secured.POST("/wallet/connect", handlers.Connect)
	secured.DELETE("/wallet/disconnect", handlers.Disconnect)
	secured.GET("/wallet/latest", handlers.Latest)

	// Bot
	secured.POST("/bot/start", handlers.Start)
	secured.POST("/bot/stop", handlers.Stop)
	secured.GET("/bot/status", handlers.Status)

	// User stats
	secured.GET("/stats", handlers.Stats)
	secured.GET("/trades/history", handlers.TradesHistory)
	secured.GET("/trades/summary", handlers.TradesSummary)

	return r
}
