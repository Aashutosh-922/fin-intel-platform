package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/infrastructure/gemini"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/infrastructure/kafka"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/infrastructure/openai"
)

func main() {
	ctx := context.Background()
	brokerEnv := os.Getenv("KAFKA_BROKERS")
	if brokerEnv == "" {
		brokerEnv = "localhost:9092"
	}
	brokers := strings.Split(brokerEnv, ",")

	analyzer := application.NewAnalyzer()
	var llm application.LLMClient
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER")))
	if provider == "" {
		provider = "gemini"
	}

	switch provider {
	case "openai":
		llm = openai.NewClient(openai.Config{
			APIKey:     os.Getenv("OPENAI_API_KEY"),
			Model:      os.Getenv("OPENAI_MODEL"),
			BaseURL:    os.Getenv("OPENAI_BASE_URL"),
			TimeoutMs:  envInt("OPENAI_TIMEOUT_MS", 4000),
			MaxRetries: envInt("OPENAI_MAX_RETRIES", 2),
		})
	case "gemini":
		llm = gemini.NewClient(gemini.Config{
			APIKey:     os.Getenv("GEMINI_API_KEY"),
			Model:      os.Getenv("GEMINI_MODEL"),
			BaseURL:    os.Getenv("GEMINI_BASE_URL"),
			TimeoutMs:  envInt("GEMINI_TIMEOUT_MS", 4000),
			MaxRetries: envInt("GEMINI_MAX_RETRIES", 2),
		})
	default:
		log.Printf("unknown AI_PROVIDER=%s; defaulting to deterministic mode fallback\n", provider)
	}

	mode := os.Getenv("AI_MODE")
	service := application.NewService(analyzer, llm, mode)

	producer := kafka.NewProducer(brokerEnv, "ai-insights")
	defer producer.Close()

	consumer := kafka.NewConsumer(
		brokers,
		"ai-service-group",
		"risk-decisions",
		service,
		producer,
	)

	log.Println("ai-service consumer started")
	consumer.Start(ctx)
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
