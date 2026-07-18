package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"dan-ai/internal/ai/provider"
	embeddingrepo "dan-ai/internal/embedding/repository"
	"dan-ai/internal/knowledge/chunk"
	"dan-ai/internal/knowledge/entity"
	"dan-ai/internal/knowledge/repository"
	promptrepo "dan-ai/internal/prompt/repository"
	"dan-ai/pkg/kafka"
	"dan-ai/pkg/milvus"
	"dan-ai/pkg/ulid"
)

const defaultChunkModel = "gemini-3.1-flash-lite"

type Processor struct {
	repo          repository.KnowledgeRepository
	aiRegistry    *provider.Registry
	milvusClient  *milvus.Client
	chunkBuilder  chunk.Builder
	promptRepo    promptrepo.Repository
	embeddingRepo embeddingrepo.Repository
}

func NewProcessor(
	repo repository.KnowledgeRepository,
	aiRegistry *provider.Registry,
	milvusClient *milvus.Client,
	chunkBuilder chunk.Builder,
	promptRepo promptrepo.Repository,
	embeddingRepo embeddingrepo.Repository,
) *Processor {
	return &Processor{
		repo:          repo,
		aiRegistry:    aiRegistry,
		milvusClient:  milvusClient,
		chunkBuilder:  chunkBuilder,
		promptRepo:    promptRepo,
		embeddingRepo: embeddingRepo,
	}
}

type knowledgeEventPayload struct {
	SourceType string `json:"source_type"`
	SourceID   string `json:"source_id"`
	PromptID   string `json:"prompt_id"`
}

func (p *Processor) ProcessEvent(ctx context.Context, event kafka.Event) error {
	log.Printf("processing knowledge event for %s %s", event.Aggregate, event.AggregateID)

	// Get all enabled embedding profiles
	enabledProfiles, err := p.embeddingRepo.ListEnabledProfiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list enabled embedding profiles: %w", err)
	}
	if len(enabledProfiles) == 0 {
		return fmt.Errorf("no enabled embedding profiles found")
	}

	// Parse payload to extract prompt_id
	var payload knowledgeEventPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		log.Printf("warning: failed to parse knowledge event payload: %v", err)
	}

	// Resolve model name, provider, and system instruction from prompt
	modelName := defaultChunkModel
	chunkProviderName := "gemini" // default provider for chunk generation
	systemInstruction := ""
	if payload.PromptID != "" {
		prompt, err := p.promptRepo.Get(ctx, payload.PromptID)
		if err == nil {
			if prompt.AIModel.Name != "" {
				modelName = prompt.AIModel.Name
			}
			if prompt.AIModel.Provider != "" {
				chunkProviderName = prompt.AIModel.Provider
			}
			systemInstruction = prompt.SystemPrompt
		}
	}

	// Fetch document by source
	doc, err := p.repo.GetDocumentBySource(ctx, event.Aggregate, event.AggregateID)
	if err != nil {
		return fmt.Errorf("failed to get knowledge document for source %s %s: %w", event.Aggregate, event.AggregateID, err)
	}

	if doc.Status == "Embedded" {
		log.Printf("Document %s is already Embedded, re-embedding...", doc.ID)
	}

	// 1. Delete old chunks from Milvus (for all enabled profiles) and PostgreSQL (once)
	for _, profile := range enabledProfiles {
		if err := p.milvusClient.DeleteVectorsByDocumentID(ctx, profile.KnowledgeCollection, doc.ID); err != nil {
			log.Printf("warning: failed to delete old vectors from milvus for doc %s in %s: %v", doc.ID, profile.KnowledgeCollection, err)
		}
	}
	if err := p.repo.DeleteChunksByDocumentID(ctx, doc.ID); err != nil {
		return fmt.Errorf("failed to delete old chunks from db: %w", err)
	}

	// 2. Generate chunks using AI Builder with resolved provider, model, and system instruction
	aiChunks, err := p.chunkBuilder.Build(ctx, chunkProviderName, modelName, systemInstruction, doc)
	if err != nil {
		return fmt.Errorf("failed to generate chunks via ai builder: %w", err)
	}

	if len(aiChunks) == 0 {
		log.Printf("no chunks generated for document %s", doc.ID)
		return nil
	}

	var chunks []entity.KnowledgeChunk
	for _, ac := range aiChunks {
		chunkID := ulid.New()
		ac.ID = chunkID
		ac.CreatedAt = time.Now()
		ac.TokenCount = 0 // can be calculated if needed
		chunks = append(chunks, ac)
	}

	// 3. Save new chunks to Postgres (once)
	if err := p.repo.CreateChunks(ctx, chunks); err != nil {
		return fmt.Errorf("failed to save chunks to db: %w", err)
	}

	// 4. Generate embeddings and upsert vectors to Milvus for each enabled embedding profile
	for _, profile := range enabledProfiles {
		var vectors []milvus.KnowledgeVector
		embeddingProvider, err := p.aiRegistry.Get(profile.Provider)
		if err != nil {
			return fmt.Errorf("failed to get embedding provider %q for profile %s: %w", profile.Provider, profile.Name, err)
		}

		for _, c := range chunks {
			embedding, err := embeddingProvider.GenerateEmbedding(ctx, profile.Model, c.Content)
			if err != nil {
				return fmt.Errorf("failed to generate embedding for chunk %s under profile %s: %w", c.ID, profile.Name, err)
			}

			vectors = append(vectors, milvus.KnowledgeVector{
				ChunkID:    c.ID,
				DocumentID: doc.ID,
				SourceType: doc.SourceType,
				SourceID:   doc.SourceID,
				Embedding:  embedding,
			})
		}

		if err := p.milvusClient.UpsertVectors(ctx, profile.KnowledgeCollection, vectors); err != nil {
			return fmt.Errorf("failed to upsert vectors to milvus for profile %s: %w", profile.Name, err)
		}
		log.Printf("successfully updated %d vector(s) in Milvus collection %s", len(vectors), profile.KnowledgeCollection)
	}

	// 5. Update document status in PostgreSQL
	now := time.Now()
	doc.Status = "Embedded"
	if len(enabledProfiles) > 0 {
		doc.EmbeddingModel = enabledProfiles[0].Model
	}
	doc.LastEmbeddedAt = &now

	if err := p.repo.UpdateDocument(ctx, doc); err != nil {
		return fmt.Errorf("failed to update document status: %w", err)
	}

	log.Printf("successfully processed and embedded document %s with %d chunks for all enabled profiles", doc.ID, len(chunks))
	return nil
}
