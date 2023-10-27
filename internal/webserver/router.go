package webserver

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/rumenvasilev/rvsecret/assets"
	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/core"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/scan/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
)

// Set various internal values used by the web interface
const (
	GithubBaseURI   = "https://raw.githubusercontent.com"
	MaximumFileSize = 153600
	GitLabBaseURL   = "https://gitlab.com"
	CspPolicy       = "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'"
	ReferrerPolicy  = "no-referrer"
	local           = "add/path/here"
	authorization   = "dW5rbm93bjp3aEB0ZXYzJCNARkRT"
	public          = "/public"
)

// Start will configure and start the webserver for graphical output and status messages
// It's a blocking call, so it should be run in a goroutine
func (e *Engine) Start() {
	if err := e.Run(e.Listener); err != nil {
		e.Logger.Fatal("Error when starting web server: %s", err)
	}
}

type Engine struct {
	*gin.Engine
	*log.Logger
	Listener string // address:port
}

// New will create an instance of the web frontend, setting the necessary parameters.
func New(cfg config.Config, state *core.State, log *log.Logger) *Engine {
	gin.SetMode(gin.ReleaseMode)
	if cfg.Global.Debug {
		gin.SetMode(gin.DebugMode)
	}

	serverRoot, err := fs.Sub(&assets.Assets, "static")
	if err != nil {
		log.Fatal(err.Error())
	}

	router := gin.New()
	router.Use(logger.SetLogger())
	router.StaticFS(public, http.FS(serverRoot))
	router.Use(secure.New(secure.Config{
		SSLRedirect:           false,
		IsDevelopment:         false,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: CspPolicy,
		ReferrerPolicy:        ReferrerPolicy,
	}))
	router.GET("/", func(c *gin.Context) {
		location := url.URL{Path: public}
		c.Redirect(http.StatusFound, location.RequestURI())
	})
	router.GET("/images/*path", rewrite{uri: "/images"}.path(router))
	router.GET("/javascripts/*path", rewrite{uri: "/javascripts"}.path(router))
	router.GET("/fonts/*path", rewrite{uri: "/fonts"}.path(router))
	router.GET("/stylesheets/*path", rewrite{uri: "/stylesheets"}.path(router))
	router.GET("/api/stats", checkAuthN, func(c *gin.Context) {
		c.JSON(200, state.Stats)
	})
	router.GET("/api/findings", checkAuthN, func(c *gin.Context) {
		c.JSON(200, state.Findings)
	})
	router.GET("/api/targets", checkAuthN, func(c *gin.Context) {
		c.JSON(200, state.Targets)
	})
	router.GET("/api/repositories", checkAuthN, func(c *gin.Context) {
		c.JSON(200, state.Repositories)
	})
	router.GET("/api/files/:owner/:repo/:commit/*path", fetch{scanType: cfg.Global.ScanType}.file)

	return &Engine{
		Listener: fmt.Sprintf("%s:%d", cfg.Global.BindAddress, cfg.Global.BindPort),
		Logger:   log,
		Engine:   router,
	}
}

func checkAuthN(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	bearer := strings.Split(authHeader, "Bearer ")
	if len(bearer) != 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if bearer[1] != authorization {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}

type rewrite struct {
	uri string
}

func (r rewrite) path(e *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.URL.Path = public + r.uri + c.Param("path")
		e.HandleContext(c)
		c.Abort()
	}
}

type fetch struct {
	scanType api.ScanType
}

// file returns a given path to a file that can be cicked on by a user
func (f fetch) file(c *gin.Context) {
	switch f.scanType {
	case api.Github:
		fileURL := fmt.Sprintf("%s/%s/%s/%s%s", GithubBaseURI, c.Param("owner"), c.Param("repo"), c.Param("commit"), c.Param("path"))
		getRemoteFile(c, fileURL)
	case api.Gitlab:
		results := util.CleanURLSpaces(c.Param("owner"), c.Param("repo"), c.Param("commit"), c.Param("path"))
		fileURL := fmt.Sprintf("%s/%s/%s/%s/%s%s", GitLabBaseURL, results[0], results[1], "/-/raw/", results[2], results[3])
		getRemoteFile(c, fileURL)
	default:
		fileURL := fmt.Sprintf("%s/%s/%s/%s%s", local, c.Param("owner"), c.Param("repo"), c.Param("commit"), c.Param("path"))
		getLocalFile(c, fileURL)
	}
}

func getRemoteFile(c *gin.Context, filepath string) {
	resp, err := http.Head(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "No content",
		})
		return
	}

	if resp.ContentLength > MaximumFileSize {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": fmt.Sprintf("File size exceeds maximum of %d bytes", MaximumFileSize),
		})
		return
	}

	resp, err = http.Get(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
		return
	}

	c.String(http.StatusOK, string(body[:]))
}

func getLocalFile(c *gin.Context, filepath string) {
	// Handle auth separately, because we don't need any for remote files
	checkAuthN(c)
	if c.IsAborted() {
		return
	}
	// defer resp.Body.Close()
	data, err := os.Open(filepath)
	//lint:ignore SA5001 ignore this
	defer data.Close() //nolint:staticcheck

	if err != nil {
		// TODO SEE ERROR AND RETURN DIFFERENT ERRORS, E.G. NOT FOUND
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
	}
	body, err := io.ReadAll(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
		return
	}
	c.String(http.StatusOK, string(body))
}
