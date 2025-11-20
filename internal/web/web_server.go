package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"tingly-box/internal/auth"
	"tingly-box/internal/config"
	"tingly-box/internal/memory"
)

// WebServer provides a simple web interface for configuration management
type WebServer struct {
	config *config.AppConfig
	logger *memory.MemoryLogger
	router *gin.Engine
}

// ProviderForm represents provider form data
type ProviderForm struct {
	Name    string `form:"name" binding:"required"`
	APIBase string `form:"api_base" binding:"required"`
	Token   string `form:"token" binding:"required"`
}

// NewWebServer creates a new web server
func NewWebServer(appConfig *config.AppConfig, logger *memory.MemoryLogger) *WebServer {
	gin.SetMode(gin.ReleaseMode)

	ws := &WebServer{
		config: appConfig,
		logger: logger,
		router: gin.New(),
	}

	ws.setupRoutes()
	return ws
}

// setupRoutes configures web server routes
func (ws *WebServer) setupRoutes() {
	// Middleware
	ws.router.Use(gin.Logger())
	ws.router.Use(gin.Recovery())

	// Static files and templates
	ws.router.LoadHTMLGlob("web/templates/*")
	ws.router.Static("/static", "./web/static")

	// API routes
	api := ws.router.Group("/api")
	{
		api.GET("/providers", ws.getProviders)
		api.POST("/providers", ws.addProvider)
		api.DELETE("/providers/:name", ws.deleteProvider)
		api.GET("/status", ws.getStatus)
		api.POST("/server/start", ws.startServer)
		api.POST("/server/stop", ws.stopServer)
		api.POST("/server/restart", ws.restartServer)
		api.GET("/token", ws.generateToken)
		api.GET("/history", ws.getHistory)
	}

	// Web page routes
	ws.router.GET("/", ws.dashboard)
	ws.router.GET("/providers", ws.providersPage)
	ws.router.GET("/server", ws.serverPage)
	ws.router.GET("/history", ws.historyPage)
}

// getRouter returns the gin router
func (ws *WebServer) GetRouter() *gin.Engine {
	return ws.router
}

// API Handlers
func (ws *WebServer) getProviders(c *gin.Context) {
	providers := ws.config.ListProviders()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    providers,
	})
}

func (ws *WebServer) addProvider(c *gin.Context) {
	var form ProviderForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := ws.config.AddProvider(form.Name, form.APIBase, form.Token); err != nil {
		if ws.logger != nil {
			ws.logger.LogAction(memory.ActionAddProvider, map[string]interface{}{
				"name":     form.Name,
				"api_base": form.APIBase,
			}, false, err.Error())
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if ws.logger != nil {
		ws.logger.LogAction(memory.ActionAddProvider, map[string]interface{}{
			"name":     form.Name,
			"api_base": form.APIBase,
		}, true, "Provider added successfully via web interface")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Provider added successfully",
	})
}

func (ws *WebServer) deleteProvider(c *gin.Context) {
	name := c.Param("name")
	if err := ws.config.DeleteProvider(name); err != nil {
		if ws.logger != nil {
			ws.logger.LogAction(memory.ActionDeleteProvider, map[string]interface{}{
				"name": name,
			}, false, err.Error())
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if ws.logger != nil {
		ws.logger.LogAction(memory.ActionDeleteProvider, map[string]interface{}{
			"name": name,
		}, true, "Provider deleted successfully via web interface")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Provider deleted successfully",
	})
}

func (ws *WebServer) getStatus(c *gin.Context) {
	providers := ws.config.ListProviders()
	enabledCount := 0
	for _, provider := range providers {
		if provider.Enabled {
			enabledCount++
		}
	}

	status := gin.H{
		"server_running": false,
		"port":           ws.config.GetServerPort(),
		"providers_total": len(providers),
		"providers_enabled": enabledCount,
	}

	// Add memory logger status if available
	if ws.logger != nil {
		currentStatus := ws.logger.GetCurrentStatus()
		status["last_updated"] = currentStatus.Timestamp.Format("2006-01-02 15:04:05")
		if currentStatus.Running {
			status["server_running"] = true
			status["uptime"] = currentStatus.Uptime
			status["request_count"] = currentStatus.RequestCount
		}

		stats := ws.logger.GetActionStats()
		status["action_stats"] = stats
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

func (ws *WebServer) startServer(c *gin.Context) {
	var req struct {
		Port int `json:"port"`
	}

	c.ShouldBindJSON(&req)
	if req.Port == 0 {
		req.Port = ws.config.GetServerPort()
	}

	// In a real implementation, this would start the actual server
	// For now, we just update the status in memory
	if ws.logger != nil {
		ws.logger.UpdateServerStatus(true, req.Port, "0s", 0)
		ws.logger.LogAction(memory.ActionStartServer, map[string]interface{}{
			"port": req.Port,
		}, true, "Server started via web interface")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Server started on port %d", req.Port),
	})
}

func (ws *WebServer) stopServer(c *gin.Context) {
	// In a real implementation, this would stop the actual server
	if ws.logger != nil {
		ws.logger.UpdateServerStatus(false, 0, "", 0)
		ws.logger.LogAction(memory.ActionStopServer, nil, true, "Server stopped via web interface")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Server stopped successfully",
	})
}

func (ws *WebServer) restartServer(c *gin.Context) {
	var req struct {
		Port int `json:"port"`
	}

	c.ShouldBindJSON(&req)
	if req.Port == 0 {
		req.Port = ws.config.GetServerPort()
	}

	// In a real implementation, this would restart the actual server
	if ws.logger != nil {
		ws.logger.UpdateServerStatus(true, req.Port, "0s", 0)
		ws.logger.LogAction(memory.ActionRestartServer, map[string]interface{}{
			"port": req.Port,
		}, true, "Server restarted via web interface")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Server restarted on port %d", req.Port),
	})
}

func (ws *WebServer) generateToken(c *gin.Context) {
	clientID := c.DefaultQuery("client_id", "web")
	jwtManager := auth.NewJWTManager(ws.config.GetJWTSecret())
	token, err := jwtManager.GenerateToken(clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if ws.logger != nil {
		ws.logger.LogAction(memory.ActionGenerateToken, map[string]interface{}{
			"client_id": clientID,
		}, true, "Token generated via web interface")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":     token,
			"client_id": clientID,
		},
	})
}

func (ws *WebServer) getHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	var history []memory.HistoryEntry
	if ws.logger != nil {
		history = ws.logger.GetHistory(limit)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// Page Handlers
func (ws *WebServer) dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Tingly Box Dashboard",
	})
}

func (ws *WebServer) providersPage(c *gin.Context) {
	c.HTML(http.StatusOK, "providers.html", gin.H{
		"title": "Providers - Tingly Box",
	})
}

func (ws *WebServer) serverPage(c *gin.Context) {
	c.HTML(http.StatusOK, "server.html", gin.H{
		"title": "Server - Tingly Box",
	})
}

func (ws *WebServer) historyPage(c *gin.Context) {
	c.HTML(http.StatusOK, "history.html", gin.H{
		"title": "History - Tingly Box",
	})
}