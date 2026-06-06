//go:build embed

package web

import (
	"bytes"
	"compress/gzip"
	"context"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

const (
	// NonceHTMLPlaceholder is the placeholder for nonce in HTML script tags
	NonceHTMLPlaceholder = "__CSP_NONCE_VALUE__"
)

//go:embed all:dist
var frontendFS embed.FS

// PublicSettingsProvider is an interface to fetch public settings
type PublicSettingsProvider interface {
	GetPublicSettingsForInjection(ctx context.Context) (any, error)
}

// FrontendServer serves the embedded frontend with settings injection
type FrontendServer struct {
	distFS      fs.FS
	fileServer  http.Handler
	baseHTML    []byte
	cache       *HTMLCache
	settings    PublicSettingsProvider
	overrideDir string // local file override directory
	// gzipAssets：启动时对可压缩静态资源预先 gzip，键为相对路径（如 assets/index-XXXX.js）。
	// 请求时若客户端支持 gzip 直接发预压缩字节，避免每请求重复压缩。
	gzipAssets map[string][]byte
}

// NewFrontendServer creates a new frontend server with settings injection
func NewFrontendServer(settingsProvider PublicSettingsProvider) (*FrontendServer, error) {
	distFS, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		return nil, err
	}

	// Read base HTML once
	file, err := distFS.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	baseHTML, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	cache := NewHTMLCache()
	cache.SetBaseHTML(baseHTML)

	return &FrontendServer{
		distFS:      distFS,
		fileServer:  http.FileServer(http.FS(distFS)),
		baseHTML:    baseHTML,
		cache:       cache,
		settings:    settingsProvider,
		overrideDir: filepath.Join("data", "public"),
		gzipAssets:  buildGzipAssets(distFS),
	}, nil
}

// buildGzipAssets 启动时遍历 dist，对可压缩文本资源预先 gzip（仅缓存确有压缩收益的）。
func buildGzipAssets(distFS fs.FS) map[string][]byte {
	out := make(map[string][]byte)
	_ = fs.WalkDir(distFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isCompressibleAsset(p) {
			return nil
		}
		data, readErr := fs.ReadFile(distFS, p)
		if readErr != nil || len(data) < 1024 { // 太小的文件压缩无意义
			return nil
		}
		var buf bytes.Buffer
		gw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		if _, werr := gw.Write(data); werr != nil {
			_ = gw.Close()
			return nil
		}
		if cerr := gw.Close(); cerr != nil {
			return nil
		}
		if buf.Len() < len(data) {
			out[p] = append([]byte(nil), buf.Bytes()...)
		}
		return nil
	})
	return out
}

// isCompressibleAsset 仅对文本类静态资源做 gzip（图片/字体已是压缩格式，跳过）。
func isCompressibleAsset(p string) bool {
	switch filepath.Ext(p) {
	case ".js", ".mjs", ".css", ".svg", ".json", ".map", ".txt", ".webmanifest":
		return true
	}
	return false
}

// staticContentType 按扩展名返回静态资源 Content-Type（gzip 直发时需显式设置）。
func staticContentType(p string) string {
	switch filepath.Ext(p) {
	case ".js", ".mjs":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".json", ".map":
		return "application/json; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".webmanifest":
		return "application/manifest+json"
	}
	return "application/octet-stream"
}

// clientAcceptsGzip 判断客户端是否接受 gzip 编码。
func clientAcceptsGzip(c *gin.Context) bool {
	return strings.Contains(c.GetHeader("Accept-Encoding"), "gzip")
}

// InvalidateCache invalidates the HTML cache (call when settings change)
func (s *FrontendServer) InvalidateCache() {
	if s != nil && s.cache != nil {
		s.cache.Invalidate()
	}
}

// Middleware returns the Gin middleware handler
func (s *FrontendServer) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API routes
		if shouldBypassEmbeddedFrontend(path) {
			c.Next()
			return
		}

		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		// For index.html or SPA routes, serve with injected settings
		if cleanPath == "index.html" || !s.fileExists(cleanPath) {
			s.serveIndexHTML(c)
			return
		}

		// 静态资源缓存策略：/assets/* 是内容哈希命名（如 index-XXXX.js），内容永不变，
		// 强缓存一年 + immutable，浏览器重复访问直接命中本地缓存（解决"每次加载都重新拉资源"）。
		// 其余静态文件（logo 等）短缓存 1 小时。
		if strings.HasPrefix(cleanPath, "assets/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			c.Header("Cache-Control", "public, max-age=3600")
		}

		// Try local override first
		if s.tryServeOverride(c, cleanPath) {
			return
		}

		// 预压缩静态资源：客户端支持 gzip 时直发启动时压缩好的字节（不阻塞、零每请求 CPU）
		if gz, ok := s.gzipAssets[cleanPath]; ok && clientAcceptsGzip(c) {
			c.Header("Content-Encoding", "gzip")
			c.Header("Vary", "Accept-Encoding")
			c.Data(http.StatusOK, staticContentType(cleanPath), gz)
			c.Abort()
			return
		}

		// Serve static files normally
		s.fileServer.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

func (s *FrontendServer) fileExists(path string) bool {
	file, err := s.distFS.Open(path)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

// tryServeOverride checks if a local override file exists and serves it.
// Files in overrideDir take precedence over embedded files.
func (s *FrontendServer) tryServeOverride(c *gin.Context, cleanPath string) bool {
	if s.overrideDir == "" {
		return false
	}
	filePath := filepath.Join(s.overrideDir, filepath.Clean("/"+cleanPath))
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return false
	}
	c.File(filePath)
	c.Abort()
	return true
}

func (s *FrontendServer) serveIndexHTML(c *gin.Context) {
	// Get nonce from context (generated by SecurityHeaders middleware)
	nonce := middleware.GetNonceFromContext(c)

	// Check cache first
	cached := s.cache.Get()
	if cached != nil {
		// Check If-None-Match for 304 response
		if match := c.GetHeader("If-None-Match"); match == cached.ETag {
			c.Status(http.StatusNotModified)
			c.Abort()
			return
		}

		// Replace nonce placeholder with actual nonce before serving
		content := replaceNoncePlaceholder(cached.Content, nonce)

		c.Header("ETag", cached.ETag)
		c.Header("Cache-Control", "no-cache") // Must revalidate
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
		c.Abort()
		return
	}

	// Cache miss - fetch settings and render
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	settings, err := s.settings.GetPublicSettingsForInjection(ctx)
	if err != nil {
		// Fallback: serve without injection
		c.Data(http.StatusOK, "text/html; charset=utf-8", s.baseHTML)
		c.Abort()
		return
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		// Fallback: serve without injection
		c.Data(http.StatusOK, "text/html; charset=utf-8", s.baseHTML)
		c.Abort()
		return
	}

	rendered := s.injectSettings(settingsJSON)
	s.cache.Set(rendered, settingsJSON)

	// Replace nonce placeholder with actual nonce before serving
	content := replaceNoncePlaceholder(rendered, nonce)

	cached = s.cache.Get()
	if cached != nil {
		c.Header("ETag", cached.ETag)
	}
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	c.Abort()
}

func (s *FrontendServer) injectSettings(settingsJSON []byte) []byte {
	// Create the script tag to inject with nonce placeholder
	// The placeholder will be replaced with actual nonce at request time
	script := []byte(`<script nonce="` + NonceHTMLPlaceholder + `">window.__APP_CONFIG__=` + string(settingsJSON) + `;</script>`)

	// Inject before </head>
	headClose := []byte("</head>")
	result := bytes.Replace(s.baseHTML, headClose, append(script, headClose...), 1)

	// Replace <title> with custom site name so the browser tab shows it immediately
	result = injectSiteTitle(result, settingsJSON)

	return result
}

// injectSiteTitle replaces the static <title> in HTML with the configured site name.
// This ensures the browser tab shows the correct title before JS executes.
func injectSiteTitle(html, settingsJSON []byte) []byte {
	var cfg struct {
		SiteName string `json:"site_name"`
	}
	if err := json.Unmarshal(settingsJSON, &cfg); err != nil || cfg.SiteName == "" {
		return html
	}

	// Find and replace the existing <title>...</title>
	titleStart := bytes.Index(html, []byte("<title>"))
	titleEnd := bytes.Index(html, []byte("</title>"))
	if titleStart == -1 || titleEnd == -1 || titleEnd <= titleStart {
		return html
	}

	newTitle := []byte("<title>" + cfg.SiteName + " - AI API Gateway</title>")
	var buf bytes.Buffer
	buf.Write(html[:titleStart])
	buf.Write(newTitle)
	buf.Write(html[titleEnd+len("</title>"):])
	return buf.Bytes()
}

// replaceNoncePlaceholder replaces the nonce placeholder with actual nonce value
func replaceNoncePlaceholder(html []byte, nonce string) []byte {
	return bytes.ReplaceAll(html, []byte(NonceHTMLPlaceholder), []byte(nonce))
}

// ServeEmbeddedFrontend returns a middleware for serving embedded frontend
// This is the legacy function for backward compatibility when no settings provider is available
func ServeEmbeddedFrontend() gin.HandlerFunc {
	distFS, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		panic("failed to get dist subdirectory: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(distFS))
	overrideDir := filepath.Join("data", "public")

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if shouldBypassEmbeddedFrontend(path) {
			c.Next()
			return
		}

		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		if file, err := distFS.Open(cleanPath); err == nil {
			_ = file.Close()
			// Try local override first
			if tryServeOverrideFile(c, overrideDir, cleanPath) {
				return
			}
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		serveIndexHTML(c, distFS)
	}
}

// tryServeOverrideFile is a standalone version of tryServeOverride for legacy usage.
func tryServeOverrideFile(c *gin.Context, overrideDir, cleanPath string) bool {
	if overrideDir == "" {
		return false
	}
	filePath := filepath.Join(overrideDir, filepath.Clean("/"+cleanPath))
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return false
	}
	c.File(filePath)
	c.Abort()
	return true
}

func shouldBypassEmbeddedFrontend(path string) bool {
	trimmed := strings.TrimSpace(path)
	return strings.HasPrefix(trimmed, "/api/") ||
		strings.HasPrefix(trimmed, "/v1/") ||
		strings.HasPrefix(trimmed, "/v1beta/") ||
		strings.HasPrefix(trimmed, "/backend-api/") ||
		strings.HasPrefix(trimmed, "/antigravity/") ||
		strings.HasPrefix(trimmed, "/setup/") ||
		trimmed == "/health" ||
		trimmed == "/responses" ||
		strings.HasPrefix(trimmed, "/responses/") ||
		strings.HasPrefix(trimmed, "/images/")
}

func serveIndexHTML(c *gin.Context, fsys fs.FS) {
	file, err := fsys.Open("index.html")
	if err != nil {
		c.String(http.StatusNotFound, "Frontend not found")
		c.Abort()
		return
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read index.html")
		c.Abort()
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	c.Abort()
}

func HasEmbeddedFrontend() bool {
	_, err := frontendFS.ReadFile("dist/index.html")
	return err == nil
}
