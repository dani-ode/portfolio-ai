// internal/chat/service/service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	aiprovider "dan-ai/internal/ai/provider"
	"dan-ai/internal/chat/entity"
	"dan-ai/internal/chat/repository"
	embeddingEntity "dan-ai/internal/embedding/entity"
	embeddingrepo "dan-ai/internal/embedding/repository"
	memoryrepo "dan-ai/internal/memory/repository"
	outboxEntity "dan-ai/internal/outbox/entity"
	outboxrepo "dan-ai/internal/outbox/repository"
	promptrepo "dan-ai/internal/prompt/repository"
	visitorsvc "dan-ai/internal/visitor/service"
	"dan-ai/pkg/config"
	"dan-ai/pkg/milvus"
	"dan-ai/pkg/ulid"
)

// Service defines the interface for Chat business operations.
type Service interface {
	// Session operations
	CreateSession(ctx context.Context, visitorID, promptID string) (*entity.ChatSession, error)
	GetSession(ctx context.Context, id string) (*entity.ChatSession, error)
	ListSessions(ctx context.Context, visitorID string) ([]entity.ChatSession, error)
	RenameSession(ctx context.Context, id, title string) (*entity.ChatSession, error)
	DeleteSession(ctx context.Context, id string) error

	// Message operations
	CreateMessage(ctx context.Context, sessionID, role, content string) (*entity.ChatMessage, error)
	ListMessages(ctx context.Context, sessionID string) ([]entity.ChatMessage, error)
	DeleteMessage(ctx context.Context, id string) error

	// Unified Chat operations
	SendChatMessage(ctx context.Context, sessionID, visitorID, promptID, content string) (*entity.ChatMessage, error)
}

type service struct {
	repo          repository.Repository
	outboxRepo    outboxrepo.Repository
	milvusClient  *milvus.Client
	aiRegistry    *aiprovider.Registry
	promptRepo    promptrepo.Repository
	visitorSvc    visitorsvc.Service
	memoryRepo    memoryrepo.Repository
	embeddingRepo embeddingrepo.Repository
}

// NewService creates a new Service instance with all required dependencies.
func NewService(
	repo repository.Repository,
	outboxRepo outboxrepo.Repository,
	milvusClient *milvus.Client,
	aiRegistry *aiprovider.Registry,
	promptRepo promptrepo.Repository,
	visitorSvc visitorsvc.Service,
	memoryRepo memoryrepo.Repository,
	embeddingRepo embeddingrepo.Repository,
) Service {
	return &service{
		repo:          repo,
		outboxRepo:    outboxRepo,
		milvusClient:  milvusClient,
		aiRegistry:    aiRegistry,
		promptRepo:    promptRepo,
		visitorSvc:    visitorSvc,
		memoryRepo:    memoryRepo,
		embeddingRepo: embeddingRepo,
	}
}

// --- Session operations ---

func (s *service) CreateSession(ctx context.Context, visitorID, promptID string) (*entity.ChatSession, error) {
	if promptID != "" {
		_, err := s.promptRepo.Get(ctx, promptID)
		if err != nil {
			promptID = ""
		}
	}
	session := &entity.ChatSession{
		ID:        ulid.New(),
		VisitorID: visitorID,
		PromptID:  promptID,
		Title:     "New Chat",
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *service) GetSession(ctx context.Context, id string) (*entity.ChatSession, error) {
	return s.repo.GetSession(ctx, id)
}

func (s *service) ListSessions(ctx context.Context, visitorID string) ([]entity.ChatSession, error) {
	return s.repo.ListSessionsByVisitor(ctx, visitorID)
}

func (s *service) RenameSession(ctx context.Context, id, title string) (*entity.ChatSession, error) {
	return s.repo.RenameSession(ctx, id, title)
}

func (s *service) DeleteSession(ctx context.Context, id string) error {
	return s.repo.DeleteSession(ctx, id)
}

// --- Message operations ---

func (s *service) CreateMessage(ctx context.Context, sessionID, role, content string) (*entity.ChatMessage, error) {
	message := &entity.ChatMessage{
		ID:        ulid.New(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Status:    "Pending",
	}

	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return nil, err
	}

	if role == "assistant" {
		session, err := s.repo.GetSession(ctx, sessionID)
		if err != nil {
			return nil, err
		}

		payload, err := json.Marshal(map[string]string{
			"visitor_id":           session.VisitorID,
			"session_id":           sessionID,
			"assistant_message_id": message.ID,
			"prompt_id":            session.PromptID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal chat.completed payload: %w", err)
		}

		outboxEvent := &outboxEntity.OutboxEvent{
			ID:          ulid.New(),
			Aggregate:   "chat_session",
			AggregateID: sessionID,
			EventType:   "chat.completed",
			Payload:     payload,
			Published:   false,
			RetryCount:  0,
			CreatedAt:   time.Now(),
		}

		if err := s.outboxRepo.CreateEvent(ctx, outboxEvent); err != nil {
			return nil, fmt.Errorf("failed to create outbox event: %w", err)
		}
	}

	return message, nil
}

func (s *service) ListMessages(ctx context.Context, sessionID string) ([]entity.ChatMessage, error) {
	return s.repo.ListMessagesBySession(ctx, sessionID)
}

func (s *service) DeleteMessage(ctx context.Context, id string) error {
	return s.repo.DeleteMessage(ctx, id)
}

func (s *service) SendChatMessage(ctx context.Context, sessionID, visitorID, promptID, content string) (*entity.ChatMessage, error) {
	// 1. Get session. If it doesn't exist, create it.
	var session *entity.ChatSession
	if sessionID != "" {
		existing, err := s.repo.GetSession(ctx, sessionID)
		if err == nil {
			session = existing
		} else if err.Error() != "record not found" {
			return nil, fmt.Errorf("failed to verify session: %w", err)
		}
	}

	if session == nil {
		// Register visitor (upsert or create if not exists)
		visitor, err := s.visitorSvc.Register(ctx, visitorID)
		if err != nil {
			return nil, fmt.Errorf("failed to register visitor: %w", err)
		}
		visitorID = visitor.ID

		// Create session
		newSession, err := s.CreateSession(ctx, visitorID, promptID)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
		session = newSession
	}

	sessionID = session.ID
	visitorID = session.VisitorID

	// Get configured embedding profile
	chatProfileName := "e5"
	cfg, err := config.Load()
	if err == nil && cfg.AI.ChatEmbeddingProfile != "" {
		chatProfileName = cfg.AI.ChatEmbeddingProfile
	}

	var profile *embeddingEntity.EmbeddingProfile
	profile, err = s.embeddingRepo.GetProfileByName(ctx, chatProfileName)
	if err != nil {
		// Fallback to first enabled profile
		enabledProfiles, err2 := s.embeddingRepo.ListEnabledProfiles(ctx)
		if err2 != nil || len(enabledProfiles) == 0 {
			return nil, fmt.Errorf("failed to get chat embedding profile: %w", err)
		}
		profile = &enabledProfiles[0]
	}

	// 2. Fetch dialogue history (before the new message is inserted)
	messages, err := s.repo.ListMessagesBySession(ctx, sessionID)
	if err != nil && err.Error() != "record not found" {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	// 3. Save the new user message to PostgreSQL
	userMsg, err := s.CreateMessage(ctx, sessionID, "user", content)
	if err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// 4. Generate Embedding for the user question using profile's provider and model
	embeddingProvider, err := s.aiRegistry.Get(profile.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding provider %q: %w", profile.Provider, err)
	}
	queryVector, err := embeddingProvider.GenerateEmbedding(ctx, profile.Model, content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// 5. Search dan_knowledge in Milvus (top 5) using profile.KnowledgeCollection
	knowledgeVectors, err := s.milvusClient.SearchKnowledge(ctx, profile.KnowledgeCollection, queryVector, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to search knowledge in milvus: %w", err)
	}

	var knowledgeChunks []string
	if len(knowledgeVectors) > 0 {
		var chunkIDs []string
		for _, kv := range knowledgeVectors {
			chunkIDs = append(chunkIDs, kv.ChunkID)
		}
		// Fetch chunks from Postgres
		chunks, err := s.repo.GetKnowledgeChunksByIDs(ctx, chunkIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load chunks content: %w", err)
		}
		for _, c := range chunks {
			knowledgeChunks = append(knowledgeChunks, fmt.Sprintf("- %s", c.Content))
		}
	}

	// 6. Search visitor_knowledge in Milvus (top 4) using profile.VisitorCollection
	visitorVectors, err := s.milvusClient.SearchVisitorMemory(ctx, profile.VisitorCollection, visitorID, queryVector, 4)
	if err != nil {
		return nil, fmt.Errorf("failed to search visitor memory: %w", err)
	}

	var visitorMemories []string
	if len(visitorVectors) > 0 {
		var memoryIDs []string
		for _, vv := range visitorVectors {
			memoryIDs = append(memoryIDs, vv.MemoryID)
		}
		// Fetch actual memory text from Postgres (since Milvus no longer stores value/text)
		memories, err := s.memoryRepo.GetMemoriesByIDs(ctx, memoryIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to load visitor memories from postgres: %w", err)
		}
		for _, m := range memories {
			visitorMemories = append(visitorMemories, fmt.Sprintf("- %s", m.MemoryText))
		}
	}

	// 7. Get system instruction from session's linked prompt
	var systemInstruction string
	var modelName string
	if session.PromptID != "" {
		p, err := s.promptRepo.Get(ctx, session.PromptID)
		if err == nil {
			systemInstruction = p.SystemPrompt
			modelName = p.AIModel.Name
		}
	}
	if systemInstruction == "" {
		// Fallback: use the first active prompt
		prompts, err := s.promptRepo.List(ctx, true)
		if err == nil && len(prompts) > 0 {
			systemInstruction = prompts[0].SystemPrompt
		}
	}

	// 8. Build dialogue history formatting (limit to last 6 messages)
	recentLimit := 6
	startIndex := 0
	if len(messages) > recentLimit {
		startIndex = len(messages) - recentLimit
	}
	recentMessages := messages[startIndex:]

	var historyLines []string
	for _, m := range recentMessages {
		roleName := "User"
		if m.Role == "assistant" {
			roleName = "Assistant"
		}
		historyLines = append(historyLines, fmt.Sprintf("%s: %s", roleName, m.Content))
	}
	historyText := strings.Join(historyLines, "\n")

	// 9. Build Prompt
	promptBuilder := strings.Builder{}
	promptBuilder.WriteString("KNOWLEDGE\n")
	if len(knowledgeChunks) > 0 {
		promptBuilder.WriteString(strings.Join(knowledgeChunks, "\n"))
	} else {
		promptBuilder.WriteString("(None)")
	}
	promptBuilder.WriteString("\n\n--------------------------------\n\n")
	promptBuilder.WriteString("VISITOR MEMORY\n")
	if len(visitorMemories) > 0 {
		promptBuilder.WriteString(strings.Join(visitorMemories, "\n"))
	} else {
		promptBuilder.WriteString("(None)")
	}
	promptBuilder.WriteString("\n\n--------------------------------\n\n")
	promptBuilder.WriteString("RECENT CHAT\n")
	if historyText != "" {
		promptBuilder.WriteString(historyText)
	} else {
		promptBuilder.WriteString("(No history yet)")
	}
	promptBuilder.WriteString("\n\n--------------------------------\n\n")
	promptBuilder.WriteString("QUESTION\n")
	promptBuilder.WriteString(content)

	promptStr := promptBuilder.String()

	// 10. Generate response using the prompt's AI model provider
	chatProviderName := "gemini" // default
	if session.PromptID != "" {
		p, err := s.promptRepo.Get(ctx, session.PromptID)
		if err == nil && p.AIModel.Provider != "" {
			chatProviderName = p.AIModel.Provider
		}
	}
	chatProvider, err := s.aiRegistry.Get(chatProviderName)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat provider %q: %w", chatProviderName, err)
	}
	aiResponse, err := chatProvider.GenerateChatResponse(ctx, modelName, systemInstruction, promptStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate chat response: %w", err)
	}

	// 11. Save Assistant reply
	assistantMsg, err := s.CreateMessage(ctx, sessionID, "assistant", aiResponse.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to save assistant response: %w", err)
	}

	// 12. Update Assistant reply with token count metadata and completion status
	assistantMsg.Model = modelName
	assistantMsg.PromptTokens = aiResponse.PromptTokens
	assistantMsg.CompletionTokens = aiResponse.CompletionTokens
	assistantMsg.LatencyMs = int32(time.Since(userMsg.CreatedAt).Milliseconds())
	assistantMsg.Status = "Completed"

	if err := s.repo.UpdateMessage(ctx, assistantMsg); err != nil {
		// Log error but don't fail the request since the message was already created and event published
		fmt.Printf("warning: failed to update assistant message metadata: %v\n", err)
	}

	return assistantMsg, nil
}
