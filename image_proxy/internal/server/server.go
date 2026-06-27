package server

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxRequestBodyBytes int64 = 1 << 20

// sizePattern matches an OpenAI-style WxH size, e.g. "1024x1024".
var sizePattern = regexp.MustCompile(`^[0-9]{1,4}x[0-9]{1,4}$`)

// Generator turns a text prompt (and optional WxH size hint) into PNG bytes.
type Generator interface {
	Generate(ctx context.Context, prompt, size string) ([]byte, error)
}

// Server wires an image Generator to HTTP handlers.
type Server struct {
	gen    Generator
	apiKey string // optional bearer token; empty disables auth
	log    *slog.Logger
}

// NewServer builds a Server. apiKey may be empty to disable authentication
// (intended for localhost-only deployments).
func NewServer(gen Generator, apiKey string, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{gen: gen, apiKey: apiKey, log: log}
}

// Handler returns the Gin engine with the routes registered.
func (s *Server) Handler() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.HandleMethodNotAllowed = true
	_ = r.SetTrustedProxies(nil)
	r.Use(gin.Recovery())
	r.GET("/healthz", s.healthz)
	r.POST("/v1/images/generations", s.generate)
	return r
}

// imageRequest mirrors the OpenAI images API request. Only prompt and size are
// acted on; n/model/response_format are accepted for compatibility but ignored
// (codex always returns a single b64_json PNG).
type imageRequest struct {
	Prompt         string `json:"prompt"`
	Size           string `json:"size"`
	N              int    `json:"n"`
	Model          string `json:"model"`
	ResponseFormat string `json:"response_format"`
}

type imageData struct {
	B64JSON string `json:"b64_json"`
}

type imageResponse struct {
	Created int64       `json:"created"`
	Data    []imageData `json:"data"`
}

func (s *Server) healthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (s *Server) generate(c *gin.Context) {
	if !s.authorized(c) {
		abortError(c, http.StatusUnauthorized, "missing or invalid api key")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
	var req imageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		abortError(c, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		abortError(c, http.StatusBadRequest, "prompt is required")
		return
	}

	size := strings.TrimSpace(req.Size)
	if size != "" && !sizePattern.MatchString(size) {
		abortError(c, http.StatusBadRequest, "size must be WxH, e.g. 1024x1024")
		return
	}

	png, err := s.gen.Generate(c.Request.Context(), prompt, size)
	if err != nil {
		s.log.Error("image generation failed", "err", err)
		abortError(c, http.StatusBadGateway, "image generation failed: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, imageResponse{
		Created: time.Now().Unix(),
		Data:    []imageData{{B64JSON: base64.StdEncoding.EncodeToString(png)}},
	})
}

func (s *Server) authorized(c *gin.Context) bool {
	if s.apiKey == "" {
		return true
	}
	return c.GetHeader("Authorization") == "Bearer "+s.apiKey
}

func abortError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{"message": msg, "type": "image_proxy_error"},
	})
}
