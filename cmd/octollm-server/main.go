package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/infinigence/octollm/pkg/composer"
	"github.com/sirupsen/logrus"

	_ "net/http/pprof"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "./config.yaml", "config file path")
	flag.Parse()

	logrus.SetLevel(logrus.DebugLevel)
	r := gin.Default()

	logrus.Infof("Using config file: %s", configFile)
	conf, err := composer.ReadConfigFile(configFile)
	if err != nil {
		logrus.WithError(err).Fatal("failed to read config file")
	}

	// Start pprof server
	if conf.PprofAddr != "" {
		go func() {
			log.Println(http.ListenAndServe(conf.PprofAddr, nil))
		}()
	}

	s := NewServer(conf)

	auth := &BearerKeyMW{}
	err = auth.UpdateFromConfig(conf)
	if err != nil {
		logrus.WithError(err).Fatal("failed to update auth from config")
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
	log.Println("listening " + conf.ListenAddr)
	log.Fatal(http.ListenAndServe(conf.ListenAddr, r))
}
