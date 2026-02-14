package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/infinigence/octollm/pkg/composer"
	"github.com/mattn/go-isatty"

	_ "net/http/pprof"
)

func main() {
	var configFile string
	var verbose bool
	var jsonLog bool
	flag.StringVar(&configFile, "c", "./config.yaml", "config file path")
	flag.BoolVar(&verbose, "v", false, "enable verbose logging")
	flag.BoolVar(&jsonLog, "json", false, "use JSON logging format")
	flag.Parse()

	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	var logHandler slog.Handler
	if isatty.IsTerminal(os.Stdout.Fd()) && !jsonLog {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level, AddSource: true})
	}
	slog.SetDefault(slog.New(logHandler))

	r := gin.Default()

	slog.Info(fmt.Sprintf("Using config file: %s", configFile))
	conf, err := composer.ReadConfigFile(configFile)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to read config file: %v", err), slog.String("config_file", configFile))
		os.Exit(1)
	}

	// Start pprof server
	if conf.PprofAddr != "" {
		go func() {
			err := http.ListenAndServe(conf.PprofAddr, nil)
			slog.Error(fmt.Sprintf("pprof server exited with error: %v", err))
		}()
	}

	s := NewServer(conf)

	auth := &BearerKeyMW{}
	err = auth.UpdateFromConfig(conf)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to update auth from config: %v", err))
		os.Exit(1)
	}

	// Register routes
	r.Use(gzip.Gzip(gzip.DefaultCompression), auth.Handle())
	r.POST("/v1/chat/completions", s.ChatCompletionsHandler())
	r.POST("/v1/completions", s.CompletionsHandler())
	r.POST("/v1/messages", s.MessagesHandler())
	r.POST("/v1/embeddings", s.EmbeddingsHandler())
	r.POST("/v1/rerank", s.RerankHandler())
	// Vertex AI / Gemini API endpoints
	// modelName includes the action suffix (e.g., "gemini-2.0-flash:generateContent")
	// This matches Google's API format where the action is part of the model identifier
	r.POST("/v1/models/:modelNameWithAction", s.VertexAIHandler())

	if conf.ListenAddr == "" {
		conf.ListenAddr = ":8080"
	}
	slog.Info(fmt.Sprintf("listening %s", conf.ListenAddr))
	err = http.ListenAndServe(conf.ListenAddr, r)
	slog.Error(fmt.Sprintf("server exited with error: %v", err))
}
