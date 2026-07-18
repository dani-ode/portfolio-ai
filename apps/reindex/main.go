package main

import (
	"context"
	"log"
	"time"

	aiClient "dan-ai/internal/ai/client"
	"dan-ai/internal/ai/provider"
	embeddingrepo "dan-ai/internal/embedding/repository"
	knowledgeEntity "dan-ai/internal/knowledge/entity"
	memoryEntity "dan-ai/internal/memory/entity"
	"dan-ai/pkg/config"
	"dan-ai/pkg/milvus"
	"dan-ai/pkg/postgres"
)

func main() {
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

	// Fetch all enabled embedding profiles
	embeddingRepo := embeddingrepo.NewPostgresRepository(db)
	enabledProfiles, err := embeddingRepo.ListEnabledProfiles(ctx)
	if err != nil {
		log.Fatalf("failed to list enabled embedding profiles: %v", err)
	}

	if len(enabledProfiles) == 0 {
		log.Fatalf("no enabled embedding profiles found in database")
	}

	log.Printf("Found %d enabled embedding profiles to re-index", len(enabledProfiles))

	// Initialize Milvus Client
	milvusCtx, milvusCancel := context.WithTimeout(ctx, 10*time.Second)
	mClient, err := milvus.NewClient(milvusCtx, cfg)
	milvusCancel()
	if err != nil {
		log.Fatalf("failed to connect to milvus: %v", err)
	}
	defer mClient.Close()

	// Initialize AI provider registry for generating embeddings
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

	// Load PostgreSQL knowledge documents and chunks (shared source of truth)
	log.Println("Loading knowledge documents and chunks from PostgreSQL...")
	var docs []knowledgeEntity.KnowledgeDocument
	if err := db.Find(&docs).Error; err != nil {
		log.Fatalf("failed to fetch knowledge documents: %v", err)
	}

	docMap := make(map[string]knowledgeEntity.KnowledgeDocument)
	for _, doc := range docs {
		docMap[doc.ID] = doc
	}

	var chunks []knowledgeEntity.KnowledgeChunk
	if err := db.Find(&chunks).Error; err != nil {
		log.Fatalf("failed to fetch knowledge chunks: %v", err)
	}

	log.Println("Loading visitor memory records from PostgreSQL...")
	var memories []memoryEntity.Memory
	if err := db.Find(&memories).Error; err != nil {
		log.Fatalf("failed to fetch visitor memories: %v", err)
	}

	// Loop over each enabled embedding profile and perform full reindexing
	for _, profile := range enabledProfiles {
		log.Printf("-----------------------------------------------------------------")
		log.Printf("Re-indexing Profile: %s (provider: %s, model: %s, dim: %d)",
			profile.Name, profile.Provider, profile.Model, profile.Dimension)

		// Initialize Collection in Milvus
		log.Printf("Initializing collections for profile %s: %s, %s", profile.Name, profile.KnowledgeCollection, profile.VisitorCollection)
		if err := mClient.InitCollection(ctx, profile.KnowledgeCollection, profile.VisitorCollection, profile.Dimension, profile.MetricType); err != nil {
			log.Fatalf("failed to init milvus collections for %s: %v", profile.Name, err)
		}

		embeddingProvider, err := aiRegistry.Get(profile.Provider)
		if err != nil {
			log.Fatalf("failed to get embedding provider %q for profile %s: %v", profile.Provider, profile.Name, err)
		}

		// --- 1. Reindex Knowledge Chunks ---
		log.Printf("Generating vectors for %d chunks under profile %s...", len(chunks), profile.Name)
		var knowledgeVectors []milvus.KnowledgeVector
		for i, c := range chunks {
			doc, ok := docMap[c.DocumentID]
			if !ok {
				log.Printf("warning: document %s not found for chunk %s, skipping", c.DocumentID, c.ID)
				continue
			}

			log.Printf("[%s][%d/%d] Generating embedding for chunk %s...", profile.Name, i+1, len(chunks), c.ID)
			embedding, err := embeddingProvider.GenerateEmbedding(ctx, profile.Model, c.Content)
			if err != nil {
				log.Fatalf("failed to generate embedding for chunk %s: %v", c.ID, err)
			}

			knowledgeVectors = append(knowledgeVectors, milvus.KnowledgeVector{
				ChunkID:    c.ID,
				DocumentID: c.DocumentID,
				SourceType: doc.SourceType,
				SourceID:   doc.SourceID,
				Embedding:  embedding,
			})
		}

		if len(knowledgeVectors) > 0 {
			log.Printf("Upserting %d knowledge vectors into Milvus collection %s...", len(knowledgeVectors), profile.KnowledgeCollection)
			batchSize := 100
			for i := 0; i < len(knowledgeVectors); i += batchSize {
				end := i + batchSize
				if end > len(knowledgeVectors) {
					end = len(knowledgeVectors)
				}
				if err := mClient.UpsertVectors(ctx, profile.KnowledgeCollection, knowledgeVectors[i:end]); err != nil {
					log.Fatalf("failed to upsert knowledge vectors for profile %s to milvus: %v", profile.Name, err)
				}
			}
		}

		// --- 2. Reindex Visitor Memories ---
		log.Printf("Generating vectors for %d visitor memories under profile %s...", len(memories), profile.Name)
		var visitorVectors []milvus.VisitorMemoryVector
		for i, mem := range memories {
			log.Printf("[%s][%d/%d] Generating embedding for visitor memory %s...", profile.Name, i+1, len(memories), mem.ID)
			embedding, err := embeddingProvider.GenerateEmbedding(ctx, profile.Model, mem.MemoryText)
			if err != nil {
				log.Fatalf("failed to generate embedding for visitor memory %s: %v", mem.ID, err)
			}

			visitorVectors = append(visitorVectors, milvus.VisitorMemoryVector{
				MemoryID:  mem.ID,
				VisitorID: mem.VisitorID,
				Embedding: embedding,
			})
		}

		if len(visitorVectors) > 0 {
			log.Printf("Upserting %d visitor memory vectors into Milvus collection %s...", len(visitorVectors), profile.VisitorCollection)
			batchSize := 100
			for i := 0; i < len(visitorVectors); i += batchSize {
				end := i + batchSize
				if end > len(visitorVectors) {
					end = len(visitorVectors)
				}
				if err := mClient.UpsertVisitorMemoryVectors(ctx, profile.VisitorCollection, visitorVectors[i:end]); err != nil {
					log.Fatalf("failed to upsert visitor memory vectors for profile %s to milvus: %v", profile.Name, err)
				}
			}
		}
	}

	// Update knowledge documents status in PostgreSQL
	log.Println("Updating knowledge documents status in PostgreSQL...")
	now := time.Now()
	// Set embedding model info to the active default or a list of enabled models
	defaultModelStr := ""
	if len(enabledProfiles) > 0 {
		defaultModelStr = enabledProfiles[0].Model
	}
	for i := range docs {
		docs[i].Status = "Embedded"
		docs[i].EmbeddingModel = defaultModelStr
		docs[i].LastEmbeddedAt = &now
		if err := db.Save(&docs[i]).Error; err != nil {
			log.Printf("warning: failed to update document %s status: %v", docs[i].ID, err)
		}
	}

	log.Println("Re-indexing completed successfully for all enabled profiles!")
}
