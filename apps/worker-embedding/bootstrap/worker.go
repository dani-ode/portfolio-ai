package bootstrap

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	aiClient "dan-ai/internal/ai/client"
	"dan-ai/internal/ai/provider"
	embeddingrepo "dan-ai/internal/embedding/repository"
	"dan-ai/internal/knowledge/chunk"
	"dan-ai/internal/knowledge/processor"
	"dan-ai/internal/knowledge/repository"
	promptrepo "dan-ai/internal/prompt/repository"
	"dan-ai/pkg/config"
	"dan-ai/pkg/kafka"
	"dan-ai/pkg/milvus"
	"dan-ai/pkg/postgres"
)

const GroupID = "dan-embedding-worker"

func RunEmbeddingWorker() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Postgres
	db, err := postgres.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	// Initialize embedding repo and active profile config
	embeddingRepo := embeddingrepo.NewPostgresRepository(db)
	enabledProfiles, err := embeddingRepo.ListEnabledProfiles(ctx)
	if err != nil {
		log.Fatalf("failed to list enabled embedding profiles: %v", err)
	}
	if len(enabledProfiles) == 0 {
		log.Fatalf("no enabled embedding profiles found")
	}

	// Initialize Milvus
	milvusCtx, milvusCancel := context.WithTimeout(ctx, 5*time.Second)
	mClient, err := milvus.NewClient(milvusCtx, cfg)
	milvusCancel()
	if err != nil {
		log.Fatalf("failed to connect to milvus: %v", err)
	}

	for _, p := range enabledProfiles {
		if err := mClient.InitCollection(ctx, p.KnowledgeCollection, p.VisitorCollection, p.Dimension, p.MetricType); err != nil {
			log.Fatalf("failed to init milvus collection for %s: %v", p.Name, err)
		}
	}


	// Initialize AI provider registry
	aiRegistry := provider.NewRegistry()

	genaiClient, err := aiClient.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to init gemini client: %v", err)
	}
	aiRegistry.Register("gemini", provider.NewGeminiProvider(genaiClient))

	if cfg.AI.OpenAIAPIKey != "" {
		aiRegistry.Register("openai", provider.NewOpenAIProvider(cfg.AI.OpenAIAPIKey))
		log.Println("OpenAI provider registered")
	}

	// Initialize chunk builder
	chunkBuilder := chunk.NewAIBuilder(aiRegistry)

	// Initialize Knowledge Processor
	repo := repository.NewPostgresKnowledgeRepository(db)
	promptRepo := promptrepo.NewPostgresRepository(db)
	proc := processor.NewProcessor(repo, aiRegistry, mClient, chunkBuilder, promptRepo, embeddingRepo)

	// Initialize Kafka Consumer
	consumer := kafka.NewConsumer(cfg, "dan.knowledge", GroupID)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("error closing consumer: %v", err)
		}
	}()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("shutting down embedding worker...")
		cancel()
	}()

	// Start consuming
	log.Println("starting embedding worker, waiting for events...")
	err = consumer.Consume(ctx, proc.ProcessEvent)
	if err != nil && err != context.Canceled {
		log.Fatalf("consumer error: %v", err)
	}
}
