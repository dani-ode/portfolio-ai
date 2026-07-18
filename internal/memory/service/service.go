package service

import (
	"context"
	"fmt"

	aiprovider "dan-ai/internal/ai/provider"
	embeddingEntity "dan-ai/internal/embedding/entity"
	embeddingrepo "dan-ai/internal/embedding/repository"
	"dan-ai/internal/memory/entity"
	"dan-ai/internal/memory/repository"
	promptrepo "dan-ai/internal/prompt/repository"
	"dan-ai/pkg/config"
	"dan-ai/pkg/milvus"
)

type Service interface {
	SaveMemories(ctx context.Context, modelName string, memories []entity.Memory) error
}

type service struct {
	repo          repository.Repository
	milvusClient  *milvus.Client
	aiRegistry    *aiprovider.Registry
	promptRepo    promptrepo.Repository
	embeddingRepo embeddingrepo.Repository
}

func NewService(repo repository.Repository, milvusClient *milvus.Client, aiRegistry *aiprovider.Registry, promptRepo promptrepo.Repository, embeddingRepo embeddingrepo.Repository) Service {
	return &service{
		repo:          repo,
		milvusClient:  milvusClient,
		aiRegistry:    aiRegistry,
		promptRepo:    promptRepo,
		embeddingRepo: embeddingRepo,
	}
}

const MergeSystemInstruction = "You are a memory consolidator. Your task is to combine two similar visitor memories into a single, cohesive memory without losing key context. Keep it short and concise (max 250 characters). Do not add metadata or introductions. Return only the consolidated memory string."

func (s *service) SaveMemories(ctx context.Context, modelName string, memories []entity.Memory) error {
	// Get all enabled embedding profiles
	enabledProfiles, err := s.embeddingRepo.ListEnabledProfiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list enabled embedding profiles: %w", err)
	}
	if len(enabledProfiles) == 0 {
		return fmt.Errorf("no enabled embedding profiles found")
	}

	// Resolve the primary chat profile name from config
	chatProfileName := "e5"
	cfg, err := config.Load()
	if err == nil && cfg.AI.ChatEmbeddingProfile != "" {
		chatProfileName = cfg.AI.ChatEmbeddingProfile
	}

	var chatProfile *embeddingEntity.EmbeddingProfile
	for i, p := range enabledProfiles {
		if p.Name == chatProfileName {
			chatProfile = &enabledProfiles[i]
			break
		}
	}
	if chatProfile == nil {
		chatProfile = &enabledProfiles[0]
	}

	// Resolve dynamic system prompt from database
	systemInstruction := MergeSystemInstruction
	allPrompts, err := s.promptRepo.List(ctx, false)
	if err == nil {
		for _, p := range allPrompts {
			if p.Name == "Memory Consolidator" {
				systemInstruction = p.SystemPrompt
				break
			}
		}
	}

	for _, memory := range memories {
		// 1. Generate query embedding for similarity search using primary chat profile
		chatEmbeddingProvider, err := s.aiRegistry.Get(chatProfile.Provider)
		if err != nil {
			return fmt.Errorf("failed to get chat embedding provider %q: %w", chatProfile.Provider, err)
		}
		chatEmbedding, err := chatEmbeddingProvider.GenerateEmbedding(ctx, chatProfile.Model, memory.MemoryText)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}

		// 2. Search Milvus for similar memories under the primary chat profile
		similarMemories, err := s.milvusClient.SearchVisitorMemory(ctx, chatProfile.VisitorCollection, memory.VisitorID, chatEmbedding, 3)
		if err != nil {
			return fmt.Errorf("failed to search similar memories: %w", err)
		}

		var targetMemory entity.Memory = memory

		// 3. Consolidate if top match is similar enough (score >= 0.85)
		if len(similarMemories) > 0 && similarMemories[0].Score >= 0.85 {
			bestMatch := similarMemories[0]

			// Fetch the existing memory text from PostgreSQL
			existingMemList, err := s.repo.GetMemoriesByIDs(ctx, []string{bestMatch.MemoryID})
			if err != nil || len(existingMemList) == 0 {
				return fmt.Errorf("failed to fetch existing memory text: %w", err)
			}
			existingMemory := existingMemList[0]

			// Call LLM to combine/merge
			prompt := fmt.Sprintf("Memory A:\n%s\n\nMemory B:\n%s\n\nCombine into one consolidated memory:", existingMemory.MemoryText, memory.MemoryText)
			chatProvider, err := s.aiRegistry.Get("gemini") // default to gemini for consolidation
			if err != nil {
				return fmt.Errorf("failed to get chat provider for consolidation: %w", err)
			}
			mergedResp, err := chatProvider.GenerateChatResponse(ctx, modelName, systemInstruction, prompt)
			if err != nil {
				return fmt.Errorf("failed to merge memories: %w", err)
			}

			targetMemory.ID = bestMatch.MemoryID
			targetMemory.MemoryText = mergedResp.Content
		}

		// 4. Save memory text to PostgreSQL (once)
		if err := s.repo.UpsertMemory(ctx, &targetMemory); err != nil {
			return fmt.Errorf("failed to save memory to postgres: %w", err)
		}

		// 5. Generate embedding and upsert to Milvus for each enabled embedding profile
		for _, profile := range enabledProfiles {
			embeddingProvider, err := s.aiRegistry.Get(profile.Provider)
			if err != nil {
				return fmt.Errorf("failed to get embedding provider %q for profile %s: %w", profile.Provider, profile.Name, err)
			}
			embedding, err := embeddingProvider.GenerateEmbedding(ctx, profile.Model, targetMemory.MemoryText)
			if err != nil {
				return fmt.Errorf("failed to generate embedding for profile %s: %w", profile.Name, err)
			}

			vector := milvus.VisitorMemoryVector{
				MemoryID:  targetMemory.ID,
				VisitorID: targetMemory.VisitorID,
				Embedding: embedding,
			}
			if err := s.milvusClient.UpsertVisitorMemoryVectors(ctx, profile.VisitorCollection, []milvus.VisitorMemoryVector{vector}); err != nil {
				return fmt.Errorf("failed to save memory vector to milvus for profile %s: %w", profile.Name, err)
			}
		}
	}
	return nil
}
