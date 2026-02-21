package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"dataray/internal/database"
	"dataray/internal/datasource"
	"dataray/internal/model"
	"dataray/internal/query"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	"github.com/uptrace/bun"
)

type Config struct {
	Host     string         `json:"Host"`
	Port     int            `json:"Port"`
	Database DatabaseConfig `json:"Database"`
	Sentry   SentryConfig   `json:"Sentry"`
}

type DatabaseConfig struct {
	Url string `json:"Url"`
}

type SentryConfig struct {
	Dsn string `json:"Dsn"`
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "f", "etc/config.toml", "the config file")
	flag.Parse()

	var c Config
	data, err := os.ReadFile(configFile)
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		os.Exit(1)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		slog.Error("Failed to parse config", "error", err)
		os.Exit(1)
	}

	db, err := database.InitDB(c.Database.Url)
	if err != nil {
		slog.Error("Failed to connect database", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrations(db); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// 使用 Gin 框架
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Sentry 中间件 (如果配置了 DSN)
	if c.Sentry.Dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              c.Sentry.Dsn,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
		}); err != nil {
			slog.Warn("Sentry initialization failed", "error", err)
		} else {
			r.Use(sentrygin.New(sentrygin.Options{
				Repanic:         true,
				WaitForDelivery: false,
			}))
			slog.Info("Sentry initialized successfully")
		}
	}

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Sentry 错误上报中间件 - 捕获非 2xx 响应
	if c.Sentry.Dsn != "" {
		r.Use(func(c *gin.Context) {
			c.Next()
			status := c.Writer.Status()
			if status >= 400 {
				msg := fmt.Sprintf("%s %s returned %d", c.Request.Method, c.Request.URL.Path, status)
				if hub := sentrygin.GetHubFromContext(c); hub != nil {
					hub.CaptureMessage(msg)
				}
			}
		})
	}

	// Request ID 中间件
	r.Use(requestIDMiddleware())

	// Access Log 中间件
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Milliseconds()
		slog.Info("access", "method", c.Request.Method, "path", c.Request.URL.Path, "duration_ms", duration, "client_ip", c.ClientIP())
	})

	// Health check
	r.GET("/health", healthHandler)
	r.GET("/sentry-test", func(ctx *gin.Context) {
		if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
			hub.CaptureMessage("Test event from DataRay")
		}
		ctx.String(http.StatusOK, "Sentry test event sent!")
	})

	// Datasource routes
	api := r.Group("/api")
	{
		// Datasources
		ds := api.Group("/datasources")
		{
			ds.GET("", listDatasources(db))
			ds.POST("", createDatasource(db))
			ds.GET("/:id", getDatasource(db))
			ds.PUT("/:id", updateDatasource(db))
			ds.DELETE("/:id", deleteDatasource(db))
			ds.POST("/test", testConnection(db))
			ds.GET("/:id/tables", getDatasourceTables(db))
			ds.GET("/:id/tables/:table/columns", getDatasourceTableColumns(db))
			ds.POST("/:id/preview", getDatasourcePreview(db))
			ds.POST("/:id/field-distribution", getFieldDistribution(db))
		}

		// Datasets
		datasets := api.Group("/datasets")
		{
			datasets.GET("", listDatasets(db))
			datasets.POST("", createDataset(db))
			datasets.GET("/:id", getDataset(db))
			datasets.DELETE("/:id", deleteDataset(db))
			datasets.GET("/:id/columns", getDatasetColumns(db))
			datasets.POST("/:id/columns", updateDatasetColumns(db))
			datasets.GET("/:id/preview", getDatasetPreview(db))
			datasets.POST("/:id/query", executeDatasetQuery(db))
		}

		// Charts
		charts := api.Group("/charts")
		{
			charts.GET("", listCharts(db))
			charts.POST("", createChart(db))
			charts.GET("/:id", getChart(db))
			charts.PUT("/:id", updateChart(db))
			charts.DELETE("/:id", deleteChart(db))
			charts.GET("/:id/data", getChartData(db))
			charts.POST("/query", executeChartQuery(db))
		}

		// Shares
		shares := api.Group("/shares")
		{
			shares.POST("", createShare(db))
			shares.GET("/:token", getShare(db))
		}
	}

	// Share view route (no /api prefix)
	r.GET("/share/:token", viewShare(db))

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	slog.Info("Server starting", "addr", addr)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited")
}

// Health handler
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Datasource handlers

func listDatasources(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := getPaginationParams(c)

		var datasources []model.Datasource
		query := db.NewSelect().Model(&datasources)
		if limit > 0 {
			query = query.Limit(limit)
		}
		if offset > 0 {
			query = query.Offset(offset)
		}
		if err := query.Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, datasources)
	}
}

func createDatasource(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name         string `json:"name"`
			Type         string `json:"type"`
			Host         string `json:"host"`
			Port         int    `json:"port"`
			DatabaseName string `json:"database_name"`
			Username     string `json:"username"`
			Password     string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Type == "" {
			req.Type = "postgresql"
		}

		ds := &model.Datasource{
			Name:         req.Name,
			Type:         req.Type,
			Host:         req.Host,
			Port:         req.Port,
			DatabaseName: req.DatabaseName,
			Username:     req.Username,
			Password:     req.Password,
		}

		if _, err := db.NewInsert().Model(ds).Returning("*").Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ds)
	}
}

func getDatasource(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		ds := &model.Datasource{ID: id}
		if err := db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ds)
	}
}

func updateDatasource(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseID(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var req struct {
			Name         string `json:"name"`
			Type         string `json:"type"`
			Host         string `json:"host"`
			Port         int    `json:"port"`
			DatabaseName string `json:"database_name"`
			Username     string `json:"username"`
			Password     string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ds := &model.Datasource{
			ID:           id,
			Name:         req.Name,
			Type:         req.Type,
			Host:         req.Host,
			Port:         req.Port,
			DatabaseName: req.DatabaseName,
			Username:     req.Username,
			Password:     req.Password,
		}

		if _, err := db.NewUpdate().Model(ds).WherePK().Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		updated := &model.Datasource{ID: id}
		if err := db.NewSelect().Model(updated).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, updated)
	}
}

func deleteDatasource(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if _, err := db.NewDelete().Model(&model.Datasource{}).Where("id = ?", id).Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

type TestConnectionReq struct {
	Type         string `json:"type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func testConnection(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TestConnectionReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Type == "" {
			req.Type = "postgresql"
		}

		driverType := datasource.DriverType(req.Type)
		driver, err := datasource.NewDriver(driverType)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported driver type: %s", req.Type)})
			return
		}

		config := datasource.ConnectionConfig{
			Host:         req.Host,
			Port:         req.Port,
			DatabaseName: req.DatabaseName,
			Username:     req.Username,
			Password:     req.Password,
		}

		if err := driver.TestConnection(c.Request.Context(), config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("connection failed: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func getDatasourceTables(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		ds := &model.Datasource{ID: id}
		if err := db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		driver, err := datasource.NewDriver(datasource.DriverType(ds.Type))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported driver type: %s", ds.Type)})
			return
		}

		config := datasource.ConnectionConfig{
			Host:         ds.Host,
			Port:         ds.Port,
			DatabaseName: ds.DatabaseName,
			Username:     ds.Username,
			Password:     ds.Password,
		}

		ctx, cancel := withDatasourceTimeout(c.Request.Context())
		defer cancel()

		conn, err := driver.Connect(ctx, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to connect: %v", err)})
			return
		}
		defer conn.Close()

		tables, err := conn.GetTables(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get tables: %v", err)})
			return
		}

		c.JSON(http.StatusOK, tables)
	}
}

func getDatasourceTableColumns(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		tableName := c.Param("table")

		ds := &model.Datasource{ID: id}
		if err := db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		driver, err := datasource.NewDriver(datasource.DriverType(ds.Type))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported driver type: %s", ds.Type)})
			return
		}

		config := datasource.ConnectionConfig{
			Host:         ds.Host,
			Port:         ds.Port,
			DatabaseName: ds.DatabaseName,
			Username:     ds.Username,
			Password:     ds.Password,
		}

		ctx, cancel := withDatasourceTimeout(c.Request.Context())
		defer cancel()

		conn, err := driver.Connect(ctx, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to connect: %v", err)})
			return
		}
		defer conn.Close()

		columns, err := conn.GetColumns(c.Request.Context(), tableName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get columns: %v", err)})
			return
		}

		result := make([]map[string]interface{}, len(columns))
		for i, col := range columns {
			result[i] = map[string]interface{}{
				"name":     col.Name,
				"dataType": col.Type,
				"comment":  col.Comment,
			}
		}
		c.JSON(http.StatusOK, result)
	}
}

func getDatasourcePreview(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		var req struct {
			TableName string `json:"table_name"`
			QuerySQL  string `json:"query_sql"`
			QueryType string `json:"query_type"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ds := &model.Datasource{ID: id}
		if err := db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		driver, err := datasource.NewDriver(datasource.DriverType(ds.Type))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported driver type: %s", ds.Type)})
			return
		}

		config := datasource.ConnectionConfig{
			Host:         ds.Host,
			Port:         ds.Port,
			DatabaseName: ds.DatabaseName,
			Username:     ds.Username,
			Password:     ds.Password,
		}

		ctx, cancel := withDatasourceTimeout(c.Request.Context())
		defer cancel()

		conn, err := driver.Connect(ctx, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to connect: %v", err)})
			return
		}
		defer conn.Close()

		var query string
		if req.QueryType == "sql" {
			query = fmt.Sprintf("%s LIMIT 10", req.QuerySQL)
		} else {
			if !isValidIdentifier(req.TableName) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid table name"})
				return
			}
			query = fmt.Sprintf("SELECT * FROM %s LIMIT 10", req.TableName)
		}

		result, err := conn.Execute(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"columns": result.Columns,
			"data":    result.Rows,
		})
	}
}

// getFieldDistribution returns the distribution of values for a specific field
func getFieldDistribution(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		var req struct {
			TableName string `json:"table_name"`
			QuerySQL  string `json:"query_sql"`
			QueryType string `json:"query_type"`
			FieldName string `json:"field_name"`
			Limit     int    `json:"limit"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.FieldName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "field_name is required"})
			return
		}

		if req.Limit <= 0 || req.Limit > 50 {
			req.Limit = 20
		}

		ds := &model.Datasource{ID: id}
		if err := db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		driver, err := datasource.NewDriver(datasource.DriverType(ds.Type))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported driver type: %s", ds.Type)})
			return
		}

		config := datasource.ConnectionConfig{
			Host:         ds.Host,
			Port:         ds.Port,
			DatabaseName: ds.DatabaseName,
			Username:     ds.Username,
			Password:     ds.Password,
		}

		ctx, cancel := withDatasourceTimeout(c.Request.Context())
		defer cancel()

		conn, err := driver.Connect(ctx, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to connect: %v", err)})
			return
		}
		defer conn.Close()

		// Build the query based on query type
		var query string
		var baseTable string

		if req.QueryType == "sql" {
			// For SQL queries, we need to use it as a subquery
			query = fmt.Sprintf("SELECT %s, COUNT(*) as _count FROM (%s) as _subquery GROUP BY %s ORDER BY _count DESC LIMIT %d",
				req.FieldName, req.QuerySQL, req.FieldName, req.Limit)
		} else {
			if !isValidIdentifier(req.TableName) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid table name"})
				return
			}
			baseTable = req.TableName
			query = fmt.Sprintf("SELECT %s, COUNT(*) as _count FROM %s GROUP BY %s ORDER BY _count DESC LIMIT %d",
				req.FieldName, baseTable, req.FieldName, req.Limit)
		}

		result, err := conn.Execute(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query: %v", err)})
			return
		}

		// Get total count for percentage calculation
		var totalCount int64
		var totalQuery string
		if req.QueryType == "sql" {
			totalQuery = fmt.Sprintf("SELECT COUNT(*) as _total FROM (%s) as _subquery", req.QuerySQL)
		} else {
			totalQuery = fmt.Sprintf("SELECT COUNT(*) as _total FROM %s", baseTable)
		}

		totalResult, err := conn.Execute(c.Request.Context(), totalQuery)
		if err == nil && len(totalResult.Rows) > 0 {
			if val, ok := totalResult.Rows[0]["_total"]; ok {
				switch v := val.(type) {
				case int64:
					totalCount = v
				case float64:
					totalCount = int64(v)
				}
			}
		}

		// Format the distribution data
		distribution := make([]map[string]interface{}, 0, len(result.Rows))
		for _, row := range result.Rows {
			value := row[req.FieldName]
			count := row["_count"]

			var countInt int64
			switch v := count.(type) {
			case int64:
				countInt = v
			case float64:
				countInt = int64(v)
			}

			var percentage float64
			if totalCount > 0 {
				percentage = float64(countInt) * 100.0 / float64(totalCount)
			}

			distribution = append(distribution, map[string]interface{}{
				"value":      value,
				"count":      countInt,
				"percentage": math.Round(percentage*100) / 100,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"field_name":   req.FieldName,
			"total_count":  totalCount,
			"unique_count": len(result.Rows),
			"distribution": distribution,
		})
	}
}

// Dataset handlers

func listDatasets(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := getPaginationParams(c)

		var datasets []model.Dataset
		query := db.NewSelect().Model(&datasets)
		if limit > 0 {
			query = query.Limit(limit)
		}
		if offset > 0 {
			query = query.Offset(offset)
		}
		if err := query.Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, datasets)
	}
}

func createDataset(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name         string `json:"name"`
			DatasourceID int    `json:"datasource_id"`
			TableName    string `json:"table_name"`
			QuerySQL     string `json:"query_sql"`
			QueryType    string `json:"query_type"`
			Mode         string `json:"mode"`
			Description  string `json:"description"`
			Tags         string `json:"tags"`
			Columns      string `json:"columns"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request body: %v", err)})
			return
		}

		ds := &model.Dataset{
			Name:         req.Name,
			DatasourceID: req.DatasourceID,
			QueryType:    req.QueryType,
			Mode:         req.Mode,
		}

		// Handle optional fields
		if req.TableName != "" {
			ds.TableName = sql.NullString{String: req.TableName, Valid: true}
		}
		if req.QuerySQL != "" {
			ds.QuerySQL = sql.NullString{String: req.QuerySQL, Valid: true}
		}
		if req.Description != "" {
			ds.Description = sql.NullString{String: req.Description, Valid: true}
		}

		// Set defaults
		if ds.QueryType == "" {
			ds.QueryType = "table"
		}
		if ds.Mode == "" {
			ds.Mode = "direct"
		}
		if ds.Tags == "" {
			ds.Tags = "[]"
		}
		if ds.QualityRules == "" {
			ds.QualityRules = "[]"
		}
		if ds.Columns == "" {
			if req.Columns != "" {
				ds.Columns = req.Columns
			} else {
				ds.Columns = "[]"
			}
		}

		// Set CreatedAt explicitly (bun doesn't auto-populate DEFAULT)
		ds.CreatedAt = sql.NullTime{Time: time.Now(), Valid: true}

		if _, err := db.NewInsert().Model(ds).Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ds)
	}
}

func getDataset(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		ds := &model.Dataset{ID: id}
		if err := db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ds)
	}
}

func deleteDataset(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if _, err := db.NewDelete().Model(&model.Dataset{}).Where("id = ?", id).Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func getDatasetColumns(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		ds := &model.Dataset{ID: id}
		if err := db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		// 如果数据库中已经有自定义列配置，优先使用
		if ds.Columns != "" && ds.Columns != "[]" {
			var savedColumns []model.DatasetColumn
			if err := json.Unmarshal([]byte(ds.Columns), &savedColumns); err == nil && len(savedColumns) > 0 {
				c.JSON(http.StatusOK, savedColumns)
				return
			}
		}

		// 从数据源获取列信息
		dsModel := &model.Datasource{ID: ds.DatasourceID}
		if err := db.NewSelect().Model(dsModel).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		driver, err := datasource.NewDriver(datasource.DriverType(dsModel.Type))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported driver type: %s", dsModel.Type)})
			return
		}

		config := datasource.ConnectionConfig{
			Host:         dsModel.Host,
			Port:         dsModel.Port,
			DatabaseName: dsModel.DatabaseName,
			Username:     dsModel.Username,
			Password:     dsModel.Password,
		}

		ctx, cancel := withDatasourceTimeout(c.Request.Context())
		defer cancel()

		conn, err := driver.Connect(ctx, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to connect: %v", err)})
			return
		}
		defer conn.Close()

		var tableName string
		if ds.QueryType == "sql" {
			tableName = fmt.Sprintf("(%s) as subq", ds.QuerySQL.String)
		} else {
			tableName = ds.TableName.String
		}

		dbColumns, err := conn.GetColumns(c.Request.Context(), tableName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get columns: %v", err)})
			return
		}

		// 转换为新的列信息结构，自动推断 role
		result := make([]model.DatasetColumn, len(dbColumns))

		// 获取类型映射器
		mapper, _ := model.NewDataTypeMapper("starrocks")

		for i, col := range dbColumns {
			stdType, typeConfig, _ := mapper.ToStandard(col.Type)

			result[i] = model.DatasetColumn{
				Name:       col.Name,
				Expr:       "`" + col.Name + "`",
				Type:       stdType,
				TypeConfig: typeConfig,
				Comment:    "",
				Role:       inferRole(string(stdType)),
			}
		}
		c.JSON(http.StatusOK, result)
	}
}

// inferRole 根据数据类型推断列的角色
func inferRole(dataType string) string {
	lowerType := ""
	if dataType != "" {
		lowerType = dataType
	}
	// 数值类型默认为 metric，其他为 dimension
	if lowerType == "int" || lowerType == "integer" ||
		lowerType == "float" || lowerType == "double" ||
		lowerType == "decimal" || lowerType == "numeric" ||
		lowerType == "real" || lowerType == "number" {
		return "metric"
	}
	return "dimension"
}

func updateDatasetColumns(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseID(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var columns []model.DatasetColumn
		if err := c.ShouldBindJSON(&columns); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		columnsJSON, err := json.Marshal(columns)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ds := &model.Dataset{ID: id}
		if err := db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ds.Columns = string(columnsJSON)
		if _, err := db.NewUpdate().Model(ds).WherePK().Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ds)
	}
}

func getDatasetPreview(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		ds := &model.Dataset{ID: id}
		if err := db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		dsModel := &model.Datasource{ID: ds.DatasourceID}
		if err := db.NewSelect().Model(dsModel).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		driver, err := datasource.NewDriver(datasource.DriverType(dsModel.Type))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported driver type: %s", dsModel.Type)})
			return
		}

		config := datasource.ConnectionConfig{
			Host:         dsModel.Host,
			Port:         dsModel.Port,
			DatabaseName: dsModel.DatabaseName,
			Username:     dsModel.Username,
			Password:     dsModel.Password,
		}

		ctx, cancel := withDatasourceTimeout(c.Request.Context())
		defer cancel()

		conn, err := driver.Connect(ctx, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to connect: %v", err)})
			return
		}
		defer conn.Close()

		var tableName string
		if ds.QueryType == "sql" {
			tableName = fmt.Sprintf("(%s) as subq", ds.QuerySQL.String)
		} else {
			tableName = ds.TableName.String
		}

		query := fmt.Sprintf("SELECT * FROM %s LIMIT 10", tableName)
		result, err := conn.Execute(c.Request.Context(), query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"columns": result.Columns,
			"data":    result.Rows,
		})
	}
}

// Chart handlers

func listCharts(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := getPaginationParams(c)

		var charts []model.Chart
		query := db.NewSelect().Model(&charts)
		if limit > 0 {
			query = query.Limit(limit)
		}
		if offset > 0 {
			query = query.Offset(offset)
		}
		if err := query.Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, charts)
	}
}

func createChart(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var chart model.Chart
		if err := c.ShouldBindJSON(&chart); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if chart.Config == "" {
			chart.Config = "{}"
		}
		if _, err := db.NewInsert().Model(&chart).Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, chart)
	}
}

func getChart(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		chart := &model.Chart{ID: id}
		if err := db.NewSelect().Model(chart).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, chart)
	}
}

type UpdateChartReq struct {
	Name      string `json:"name"`
	ChartType string `json:"chart_type"`
	Config    string `json:"config"`
}

func updateChart(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := parseID(c, "id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var req UpdateChartReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		chart := &model.Chart{ID: id}
		if err := db.NewSelect().Model(chart).WherePK().Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		chart.Name = req.Name
		chart.ChartType = req.ChartType
		chart.Config = req.Config
		if _, err := db.NewUpdate().Model(chart).WherePK().Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, chart)
	}
}

func deleteChart(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if _, err := db.NewDelete().Model(&model.Chart{}).Where("id = ?", id).Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func getChartData(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		ctx := c.Request.Context()

		chart := &model.Chart{ID: id}
		if err := db.NewSelect().Model(chart).WherePK().Scan(ctx); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		dataset := &model.Dataset{ID: chart.DatasetID}
		if err := db.NewSelect().Model(dataset).WherePK().Scan(ctx); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		dsModel := &model.Datasource{ID: dataset.DatasourceID}
		if err := db.NewSelect().Model(dsModel).WherePK().Scan(ctx); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		driver, err := datasource.NewDriver(datasource.DriverType(dsModel.Type))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported driver type: %s", dsModel.Type)})
			return
		}

		config := datasource.ConnectionConfig{
			Host:         dsModel.Host,
			Port:         dsModel.Port,
			DatabaseName: dsModel.DatabaseName,
			Username:     dsModel.Username,
			Password:     dsModel.Password,
		}

		ctx, cancel := withDatasourceTimeout(ctx)
		defer cancel()

		conn, err := driver.Connect(ctx, config)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("invalid connection: %v", err)})
			return
		}
		defer conn.Close()

		var chartConfig map[string]interface{}
		json.Unmarshal([]byte(chart.Config), &chartConfig)

		sql, err := buildSQL(dataset, chartConfig, dsModel.Type)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := conn.Execute(ctx, sql)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("query failed: %v", err)})
			return
		}

		c.JSON(http.StatusOK, result.Rows)
	}
}

func parseDatasetColumns(columnsJSON string) ([]model.DatasetColumn, error) {
	if columnsJSON == "" || columnsJSON == "[]" {
		return nil, nil
	}
	var columns []model.DatasetColumn
	if err := json.Unmarshal([]byte(columnsJSON), &columns); err != nil {
		return nil, fmt.Errorf("failed to parse columns: %v", err)
	}
	return columns, nil
}

func resolveExpr(expr string, columns []model.DatasetColumn) (string, error) {
	if expr == "" {
		return "", nil
	}

	result := expr

	for i := 0; i < len(columns); i++ {
		col := columns[i]
		if col.Name == "" || col.Expr == "" {
			continue
		}

		pattern := "[" + col.Name + "]"
		if strings.Contains(result, pattern) {
			result = strings.ReplaceAll(result, pattern, col.Expr)
		}
	}

	return result, nil
}

func buildSQL(ds *model.Dataset, config map[string]interface{}, driverType string) (string, error) {
	columns, err := parseDatasetColumns(ds.Columns)
	if err != nil {
		return "", err
	}

	xAxis, _ := config["xAxis"].(string)
	yAxis, _ := config["yAxis"].(string)
	chartType, _ := config["chartType"].(string)

	if xAxis == "" {
		return "", fmt.Errorf("xAxis is required")
	}

	xAxisExpr, err := resolveExpr(xAxis, columns)
	if err != nil {
		return "", fmt.Errorf("invalid xAxis: %v", err)
	}

	yAxisExpr := ""
	if yAxis != "" {
		yAxisExpr, err = resolveExpr(yAxis, columns)
		if err != nil {
			return "", fmt.Errorf("invalid yAxis: %v", err)
		}
	}

	var baseQuery string
	if ds.QueryType == "sql" && ds.QuerySQL.Valid {
		baseQuery = fmt.Sprintf("(%s) AS subq", ds.QuerySQL.String)
	} else if ds.TableName.Valid {
		if !isValidIdentifier(ds.TableName.String) {
			return "", fmt.Errorf("invalid table name")
		}
		baseQuery = ds.TableName.String
	} else {
		return "", fmt.Errorf("no table or query defined in dataset")
	}

	selectCols := xAxisExpr
	if yAxis != "" {
		if chartType == "pie" {
			selectCols = fmt.Sprintf("%s, SUM(%s) as value", xAxisExpr, yAxisExpr)
		} else {
			selectCols = fmt.Sprintf("%s, %s", xAxisExpr, yAxisExpr)
		}
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", selectCols, baseQuery)

	if chartType == "pie" || chartType == "bar" {
		sql += fmt.Sprintf(" GROUP BY %s", xAxisExpr)
	}

	sql += fmt.Sprintf(" ORDER BY %s", xAxisExpr)
	sql += " LIMIT 1000"

	return sql, nil
}

// Share handlers

type CreateShareReq struct {
	ChartID int `json:"chart_id"`
}

func createShare(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateShareReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token := generateToken()
		share := &model.Share{
			Token:   token,
			ChartID: req.ChartID,
		}

		if _, err := db.NewInsert().Model(share).Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, share)
	}
}

func getShare(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")

		share := &model.Share{}
		if err := db.NewSelect().Model(share).Where("token = ?", token).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			return
		}

		c.JSON(http.StatusOK, share)
	}
}

func viewShare(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")

		share := &model.Share{}
		if err := db.NewSelect().Model(share).Where("token = ?", token).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			return
		}

		if share.ExpiresAt.Valid && share.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusGone, gin.H{"error": "share link has expired"})
			return
		}

		c.Redirect(http.StatusFound, "/#/share/"+token)
	}
}

func generateToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		now := time.Now().UnixNano()
		return fmt.Sprintf("%d-%d", now, now%100000000)
	}
	return fmt.Sprintf("%s-%x", hex.EncodeToString(b[:8]), b[8:])
}

// parseID validates and parses an ID from path parameter
func parseID(c *gin.Context, paramName string) (int, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, fmt.Errorf("missing %s parameter", paramName)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s parameter: must be a positive integer", paramName)
	}
	return id, nil
}

// isValidIdentifier checks if a name is a valid SQL identifier
// Valid identifiers start with a letter or underscore and contain only
// alphanumeric characters and underscores
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	// Must start with letter or underscore
	if !unicode.IsLetter(rune(name[0])) && name[0] != '_' {
		return false
	}
	// Must contain only alphanumeric and underscore
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
			return false
		}
	}
	return true
}

const datasourceTimeout = 30 * time.Second

func withDatasourceTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, datasourceTimeout)
}

func getPaginationParams(c *gin.Context) (limit, offset int) {
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ = strconv.Atoi(limitStr)
	offset, _ = strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return
}

type datasourceConnection struct {
	ds   *model.Datasource
	conn datasource.Connection
}

func getDatasourceConnection(db *bun.DB, ctx context.Context, datasourceID int) (*datasourceConnection, error) {
	ds := &model.Datasource{ID: datasourceID}
	if err := db.NewSelect().Model(ds).WherePK().Scan(ctx); err != nil {
		return nil, fmt.Errorf("datasource not found: %v", err)
	}

	driver, err := datasource.NewDriver(datasource.DriverType(ds.Type))
	if err != nil {
		return nil, fmt.Errorf("unsupported driver type: %s", ds.Type)
	}

	config := datasource.ConnectionConfig{
		Host:         ds.Host,
		Port:         ds.Port,
		DatabaseName: ds.DatabaseName,
		Username:     ds.Username,
		Password:     ds.Password,
	}

	conn, err := driver.Connect(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &datasourceConnection{ds: ds, conn: conn}, nil
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func bindAndValidate(c *gin.Context, dest interface{}) []ValidationError {
	var errors []ValidationError

	if err := c.ShouldBindJSON(dest); err != nil {
		errors = append(errors, ValidationError{
			Field:   "body",
			Message: err.Error(),
		})
		return errors
	}

	return validateStruct(dest)
}

func validateStruct(s interface{}) []ValidationError {
	var errors []ValidationError

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return errors
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		fieldName := jsonTag
		if idx := strings.Index(jsonTag, ","); idx > 0 {
			fieldName = jsonTag[:idx]
		}

		if field.Kind() == reflect.String {
			value := field.String()
			if required := fieldType.Tag.Get("required"); required == "true" && value == "" {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: "is required",
				})
			}
		}

		if field.Kind() == reflect.Int && field.Int() == 0 {
			if required := fieldType.Tag.Get("required"); required == "true" {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: "is required",
				})
			}
		}
	}

	return errors
}

func respondValidationError(c *gin.Context, errors []ValidationError) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":  "validation failed",
		"errors": errors,
	})
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			b := make([]byte, 8)
			rand.Read(b)
			requestID = fmt.Sprintf("%x", b)
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func getRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		return id.(string)
	}
	return ""
}

// Query Config types for dataset query

type QueryConfig struct {
	DimensionGroups []FieldGroup `json:"dimension_groups"`
	MetricGroups    []FieldGroup `json:"metric_groups"`
	Filters         []Filter     `json:"filters"`
	Sort            *SortConfig  `json:"sort"`
	Limit           int          `json:"limit"`
}

type FieldGroup struct {
	ID     string   `json:"id"`
	Fields []string `json:"fields"`
	Alias  string   `json:"alias,omitempty"`
}

type Filter struct {
	ID       string      `json:"id"`
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	ValueEnd interface{} `json:"value_end,omitempty"`
	Logic    string      `json:"logic"`
}

type SortConfig struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

// executeDatasetQuery handles POST /api/datasets/:id/query
func executeDatasetQuery(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dataset id"})
			return
		}

		var queryConfig QueryConfig
		if err := c.ShouldBindJSON(&queryConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var dataset model.Dataset
		if err := db.NewSelect().Model(&dataset).Where("id = ?", datasetID).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "dataset not found"})
			return
		}

		var ds model.Datasource
		if err := db.NewSelect().Model(&ds).Where("id = ?", dataset.DatasourceID).Scan(c.Request.Context()); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "datasource not found"})
			return
		}

		driverType := datasource.DriverType(ds.Type)
		driver, err := datasource.NewDriver(driverType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create driver: " + err.Error()})
			return
		}

		connConfig := datasource.ConnectionConfig{
			Host:         ds.Host,
			Port:         ds.Port,
			DatabaseName: ds.DatabaseName,
			Username:     ds.Username,
			Password:     ds.Password,
		}

		conn, err := driver.Connect(c.Request.Context(), connConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect: " + err.Error()})
			return
		}
		defer conn.Close()

		baseSQL := dataset.QuerySQL.String
		if baseSQL == "" {
			baseSQL = "SELECT * FROM " + dataset.TableName.String
		}

		querySQL, err := buildQuerySQL(baseSQL, queryConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build query: " + err.Error()})
			return
		}

		result, err := conn.Execute(c.Request.Context(), querySQL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, result.Rows)
	}
}

func buildQuerySQL(baseSQL string, config QueryConfig) (string, error) {
	var sb strings.Builder
	sb.WriteString("SELECT ")

	if len(config.MetricGroups) == 0 {
		sb.WriteString("*")
	} else {
		var selectFields []string
		for _, mg := range config.MetricGroups {
			for _, field := range mg.Fields {
				selectFields = append(selectFields, field)
			}
		}
		if len(config.DimensionGroups) > 0 {
			for _, dg := range config.DimensionGroups {
				for _, field := range dg.Fields {
					selectFields = append(selectFields, field)
				}
			}
		}
		if len(selectFields) == 0 {
			sb.WriteString("*")
		} else {
			sb.WriteString(strings.Join(selectFields, ", "))
		}
	}

	sb.WriteString(" FROM (")
	sb.WriteString(baseSQL)
	sb.WriteString(") AS t")

	if len(config.Filters) > 0 {
		sb.WriteString(" WHERE ")
		for i, f := range config.Filters {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(buildFilterSQL(f))
		}
	}

	if len(config.DimensionGroups) > 0 {
		sb.WriteString(" GROUP BY ")
		var groupByFields []string
		for _, dg := range config.DimensionGroups {
			for _, field := range dg.Fields {
				groupByFields = append(groupByFields, field)
			}
		}
		sb.WriteString(strings.Join(groupByFields, ", "))
	}

	if config.Sort != nil {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(config.Sort.Field)
		sb.WriteString(" ")
		sb.WriteString(config.Sort.Order)
	}

	if config.Limit > 0 {
		sb.WriteString(" LIMIT ")
		sb.WriteString(strconv.Itoa(config.Limit))
	}

	return sb.String(), nil
}

func buildFilterSQL(f Filter) string {
	field := f.Field
	switch f.Operator {
	case "eq":
		return fmt.Sprintf("%s = '%v'", field, f.Value)
	case "neq":
		return fmt.Sprintf("%s <> '%v'", field, f.Value)
	case "gt":
		return fmt.Sprintf("%s > '%v'", field, f.Value)
	case "gte":
		return fmt.Sprintf("%s >= '%v'", field, f.Value)
	case "lt":
		return fmt.Sprintf("%s < '%v'", field, f.Value)
	case "lte":
		return fmt.Sprintf("%s <= '%v'", field, f.Value)
	case "like":
		return fmt.Sprintf("%s LIKE '%%%v%%'", field, f.Value)
	case "in":
		if vals, ok := f.Value.([]interface{}); ok {
			var strVals []string
			for _, v := range vals {
				strVals = append(strVals, fmt.Sprintf("'%v'", v))
			}
			return fmt.Sprintf("%s IN (%s)", field, strings.Join(strVals, ", "))
		}
		return fmt.Sprintf("%s IN ('%v')", field, f.Value)
	case "between":
		return fmt.Sprintf("%s BETWEEN '%v' AND '%v'", field, f.Value, f.ValueEnd)
	case "isNull":
		return fmt.Sprintf("%s IS NULL", field)
	case "isNotNull":
		return fmt.Sprintf("%s IS NOT NULL", field)
	default:
		return fmt.Sprintf("%s = '%v'", field, f.Value)
	}
}

// executeChartQuery 处理 POST /api/charts/query
func executeChartQuery(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req query.ChartQueryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx := c.Request.Context()

		// 获取数据集
		dataset := &model.Dataset{ID: req.DatasetID}
		if err := db.NewSelect().Model(dataset).WherePK().Scan(ctx); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "dataset not found"})
			return
		}

		// 获取数据源
		ds := &model.Datasource{ID: dataset.DatasourceID}
		if err := db.NewSelect().Model(ds).WherePK().Scan(ctx); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "datasource not found"})
			return
		}

		// 创建驱动和连接
		driver, err := datasource.NewDriver(datasource.DriverType(ds.Type))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create driver: " + err.Error()})
			return
		}

		connConfig := datasource.ConnectionConfig{
			Host:         ds.Host,
			Port:         ds.Port,
			DatabaseName: ds.DatabaseName,
			Username:     ds.Username,
			Password:     ds.Password,
		}

		conn, err := driver.Connect(ctx, connConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect: " + err.Error()})
			return
		}
		defer conn.Close()

		// 创建执行器
		executor := query.NewExecutor(conn, dataset, ds)

		// 执行查询
		result, err := executor.Execute(ctx, &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
