// internal/chat/grpc/handler_test.go
package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"portfolio-ai/internal/chat/entity"
	visitorentity "portfolio-ai/internal/visitor/entity"
	visitorsvc "portfolio-ai/internal/visitor/service"
	pb "portfolio-ai/proto/chat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Manual mocks
// ---------------------------------------------------------------------------

// mockChatService implements service.Service.
type mockChatService struct {
	getSessionFn    func(ctx context.Context, id string) (*entity.ChatSession, error)
	createSessionFn func(ctx context.Context, visitorID, promptID string) (*entity.ChatSession, error)
	listMessagesFn  func(ctx context.Context, sessionID string) ([]entity.ChatMessage, error)
	createMessageFn func(ctx context.Context, sessionID, role, content string) (*entity.ChatMessage, error)
}

func (m *mockChatService) GetSession(ctx context.Context, id string) (*entity.ChatSession, error) {
	if m.getSessionFn != nil {
		return m.getSessionFn(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockChatService) CreateSession(ctx context.Context, visitorID, promptID string) (*entity.ChatSession, error) {
	if m.createSessionFn != nil {
		return m.createSessionFn(ctx, visitorID, promptID)
	}
	return &entity.ChatSession{ID: "new-session-id", VisitorID: visitorID, PromptID: promptID}, nil
}

func (m *mockChatService) ListMessages(ctx context.Context, sessionID string) ([]entity.ChatMessage, error) {
	if m.listMessagesFn != nil {
		return m.listMessagesFn(ctx, sessionID)
	}
	return nil, nil
}

func (m *mockChatService) CreateMessage(ctx context.Context, sessionID, role, content string) (*entity.ChatMessage, error) {
	if m.createMessageFn != nil {
		return m.createMessageFn(ctx, sessionID, role, content)
	}
	return &entity.ChatMessage{ID: "msg-id", SessionID: sessionID, Role: role, Content: content}, nil
}

func (m *mockChatService) ListSessions(ctx context.Context, visitorID string) ([]entity.ChatSession, error) {
	return nil, nil
}
func (m *mockChatService) RenameSession(ctx context.Context, id, title string) (*entity.ChatSession, error) {
	return nil, nil
}
func (m *mockChatService) DeleteSession(ctx context.Context, id string) error { return nil }
func (m *mockChatService) DeleteMessage(ctx context.Context, id string) error { return nil }

// mockVisitorService implements visitorsvc.Service.
type mockVisitorService struct {
	registerFn func(ctx context.Context, visitorID string) (*visitorentity.Visitor, error)
}

func (m *mockVisitorService) Register(ctx context.Context, visitorID string) (*visitorentity.Visitor, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, visitorID)
	}
	return &visitorentity.Visitor{ID: "visitor-id"}, nil
}

func (m *mockVisitorService) Get(ctx context.Context, id string) (*visitorentity.Visitor, error) {
	return nil, nil
}

// compile-time interface checks
var _ visitorsvc.Service = (*mockVisitorService)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestHandler(svc *mockChatService, vsvc visitorsvc.Service) *Handler {
	return NewHandler(svc, vsvc)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestSendChatMessage_ExistingSession verifies that when a valid session_id is
// provided and GetSession succeeds, the handler reuses the session without
// creating a new one or registering a visitor.
func TestSendChatMessage_ExistingSession(t *testing.T) {
	ctx := context.Background()

	existingSession := &entity.ChatSession{
		ID:        "session-123",
		VisitorID: "visitor-456",
		PromptID:  "prompt-789",
	}

	svc := &mockChatService{
		getSessionFn: func(_ context.Context, id string) (*entity.ChatSession, error) {
			assert.Equal(t, "session-123", id)
			return existingSession, nil
		},
		listMessagesFn: func(_ context.Context, sessionID string) ([]entity.ChatMessage, error) {
			assert.Equal(t, "session-123", sessionID)
			return []entity.ChatMessage{
				{ID: "m1", Role: "user", Content: "hello", CreatedAt: time.Now()},
				{ID: "m2", Role: "assistant", Content: "hi", CreatedAt: time.Now()},
			}, nil
		},
		createMessageFn: func(_ context.Context, sessionID, role, content string) (*entity.ChatMessage, error) {
			assert.Equal(t, "session-123", sessionID)
			assert.Equal(t, "user", role)
			assert.Equal(t, "new message", content)
			return &entity.ChatMessage{ID: "m3"}, nil
		},
	}

	// visitorSvc must NOT be called when session already exists.
	vsvc := &mockVisitorService{
		registerFn: func(_ context.Context, _ string) (*visitorentity.Visitor, error) {
			t.Fatal("Register should not be called when session already exists")
			return nil, nil
		},
	}

	h := newTestHandler(svc, vsvc)
	resp, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		SessionId: "session-123",
		Content:   "new message",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "session-123", resp.GetSessionId())
	assert.Equal(t, "visitor-456", resp.GetVisitorId())
	assert.Equal(t, "new message", resp.GetContent())
	// 2 prior messages → both returned (≤ 3)
	assert.Len(t, resp.GetRecentMessages(), 2)
}

// TestSendChatMessage_ExistingSession_InheritsVisitorAndPrompt verifies that
// if visitor_id and prompt_id are empty in the request, they are filled from
// the existing session.
func TestSendChatMessage_ExistingSession_InheritsVisitorAndPrompt(t *testing.T) {
	ctx := context.Background()

	svc := &mockChatService{
		getSessionFn: func(_ context.Context, _ string) (*entity.ChatSession, error) {
			return &entity.ChatSession{
				ID:        "sess-1",
				VisitorID: "vis-from-session",
				PromptID:  "prompt-from-session",
			}, nil
		},
	}

	vsvc := &mockVisitorService{
		registerFn: func(_ context.Context, _ string) (*visitorentity.Visitor, error) {
			t.Fatal("Register should not be called when session already exists")
			return nil, nil
		},
	}

	h := newTestHandler(svc, vsvc)
	resp, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		SessionId: "sess-1",
		Content:   "hi",
	})

	require.NoError(t, err)
	assert.Equal(t, "vis-from-session", resp.GetVisitorId())
}

// TestSendChatMessage_NewSession_NoSessionID verifies that when no session_id
// is provided, the handler registers a visitor and creates a new session.
func TestSendChatMessage_NewSession_NoSessionID(t *testing.T) {
	ctx := context.Background()

	registerCalled := false
	createSessionCalled := false

	vsvc := &mockVisitorService{
		registerFn: func(_ context.Context, visitorID string) (*visitorentity.Visitor, error) {
			registerCalled = true
			assert.Equal(t, "", visitorID) // no visitor ID provided
			return &visitorentity.Visitor{ID: "new-visitor-id"}, nil
		},
	}

	svc := &mockChatService{
		createSessionFn: func(_ context.Context, visitorID, promptID string) (*entity.ChatSession, error) {
			createSessionCalled = true
			assert.Equal(t, "new-visitor-id", visitorID)
			return &entity.ChatSession{ID: "new-sess-id", VisitorID: "new-visitor-id"}, nil
		},
	}

	h := newTestHandler(svc, vsvc)
	resp, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		Content: "hello world",
	})

	require.NoError(t, err)
	assert.True(t, registerCalled, "Register must be called")
	assert.True(t, createSessionCalled, "CreateSession must be called")
	assert.Equal(t, "new-sess-id", resp.GetSessionId())
	assert.Equal(t, "new-visitor-id", resp.GetVisitorId())
	assert.Equal(t, "hello world", resp.GetContent())
}

// TestSendChatMessage_SessionNotFound_CreatesNew verifies that when session_id
// is provided but GetSession returns ErrRecordNotFound, a new session is
// created (falls through to the new-session path).
func TestSendChatMessage_SessionNotFound_CreatesNew(t *testing.T) {
	ctx := context.Background()

	svc := &mockChatService{
		getSessionFn: func(_ context.Context, _ string) (*entity.ChatSession, error) {
			return nil, gorm.ErrRecordNotFound
		},
		createSessionFn: func(_ context.Context, visitorID, _ string) (*entity.ChatSession, error) {
			return &entity.ChatSession{ID: "created-sess", VisitorID: visitorID}, nil
		},
	}

	vsvc := &mockVisitorService{
		registerFn: func(_ context.Context, _ string) (*visitorentity.Visitor, error) {
			return &visitorentity.Visitor{ID: "registered-visitor"}, nil
		},
	}

	h := newTestHandler(svc, vsvc)
	resp, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		SessionId: "ghost-session",
		Content:   "anyone there?",
	})

	require.NoError(t, err)
	assert.Equal(t, "created-sess", resp.GetSessionId())
}

// TestSendChatMessage_RecentMessages_SlicesLastThree verifies that only the
// last 3 messages (out of more than 3 prior messages) are returned.
func TestSendChatMessage_RecentMessages_SlicesLastThree(t *testing.T) {
	ctx := context.Background()

	msgs := []entity.ChatMessage{
		{ID: "m1", Role: "user", Content: "msg1", CreatedAt: time.Now()},
		{ID: "m2", Role: "assistant", Content: "msg2", CreatedAt: time.Now()},
		{ID: "m3", Role: "user", Content: "msg3", CreatedAt: time.Now()},
		{ID: "m4", Role: "assistant", Content: "msg4", CreatedAt: time.Now()},
		{ID: "m5", Role: "user", Content: "msg5", CreatedAt: time.Now()},
	}

	svc := &mockChatService{
		getSessionFn: func(_ context.Context, _ string) (*entity.ChatSession, error) {
			return &entity.ChatSession{ID: "sess", VisitorID: "vis"}, nil
		},
		listMessagesFn: func(_ context.Context, _ string) ([]entity.ChatMessage, error) {
			return msgs, nil
		},
	}

	h := newTestHandler(svc, &mockVisitorService{})
	resp, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		SessionId: "sess",
		Content:   "latest message",
	})

	require.NoError(t, err)
	// 5 prior messages → only the last 3 are returned
	assert.Len(t, resp.GetRecentMessages(), 3)
	assert.Equal(t, "msg3", resp.GetRecentMessages()[0].GetContent())
	assert.Equal(t, "msg4", resp.GetRecentMessages()[1].GetContent())
	assert.Equal(t, "msg5", resp.GetRecentMessages()[2].GetContent())
}

// TestSendChatMessage_GetSession_InternalError verifies that an unexpected
// error from GetSession propagates as an Internal gRPC status.
func TestSendChatMessage_GetSession_InternalError(t *testing.T) {
	ctx := context.Background()

	svc := &mockChatService{
		getSessionFn: func(_ context.Context, _ string) (*entity.ChatSession, error) {
			return nil, errors.New("db timeout")
		},
	}

	h := newTestHandler(svc, &mockVisitorService{})
	_, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		SessionId: "sess-bad",
		Content:   "hello",
	})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to verify session")
}

// TestSendChatMessage_RegisterVisitor_Error verifies that a failure in
// visitor registration returns an Internal gRPC error.
func TestSendChatMessage_RegisterVisitor_Error(t *testing.T) {
	ctx := context.Background()

	vsvc := &mockVisitorService{
		registerFn: func(_ context.Context, _ string) (*visitorentity.Visitor, error) {
			return nil, errors.New("registry unavailable")
		},
	}

	h := newTestHandler(&mockChatService{}, vsvc)
	_, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		Content: "hello",
	})

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to register visitor")
}

// TestSendChatMessage_CreateSession_Error verifies that a failure in session
// creation returns an Internal gRPC error.
func TestSendChatMessage_CreateSession_Error(t *testing.T) {
	ctx := context.Background()

	svc := &mockChatService{
		createSessionFn: func(_ context.Context, _, _ string) (*entity.ChatSession, error) {
			return nil, errors.New("session store down")
		},
	}

	h := newTestHandler(svc, &mockVisitorService{})
	_, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		Content: "hello",
	})

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to create session")
}

// TestSendChatMessage_ListMessages_InternalError verifies that a non-NotFound
// error from ListMessages returns an Internal gRPC error.
func TestSendChatMessage_ListMessages_InternalError(t *testing.T) {
	ctx := context.Background()

	svc := &mockChatService{
		getSessionFn: func(_ context.Context, _ string) (*entity.ChatSession, error) {
			return &entity.ChatSession{ID: "sess", VisitorID: "vis"}, nil
		},
		listMessagesFn: func(_ context.Context, _ string) ([]entity.ChatMessage, error) {
			return nil, errors.New("query failed")
		},
	}

	h := newTestHandler(svc, &mockVisitorService{})
	_, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		SessionId: "sess",
		Content:   "hi",
	})

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to list messages")
}

// TestSendChatMessage_CreateMessage_Error verifies that a failure in message
// creation returns an Internal gRPC error.
func TestSendChatMessage_CreateMessage_Error(t *testing.T) {
	ctx := context.Background()

	svc := &mockChatService{
		getSessionFn: func(_ context.Context, _ string) (*entity.ChatSession, error) {
			return &entity.ChatSession{ID: "sess", VisitorID: "vis"}, nil
		},
		createMessageFn: func(_ context.Context, _, _, _ string) (*entity.ChatMessage, error) {
			return nil, errors.New("disk full")
		},
	}

	h := newTestHandler(svc, &mockVisitorService{})
	_, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		SessionId: "sess",
		Content:   "hi",
	})

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "failed to create message")
}

// TestSendChatMessage_ListMessages_RecordNotFound_OK verifies that a
// gorm.ErrRecordNotFound from ListMessages is treated as an empty list (not
// an error), and the handler still succeeds.
func TestSendChatMessage_ListMessages_RecordNotFound_OK(t *testing.T) {
	ctx := context.Background()

	svc := &mockChatService{
		getSessionFn: func(_ context.Context, _ string) (*entity.ChatSession, error) {
			return &entity.ChatSession{ID: "sess", VisitorID: "vis"}, nil
		},
		listMessagesFn: func(_ context.Context, _ string) ([]entity.ChatMessage, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	h := newTestHandler(svc, &mockVisitorService{})
	resp, err := h.SendChatMessage(ctx, &pb.SendChatMessageRequest{
		SessionId: "sess",
		Content:   "first message ever",
	})

	require.NoError(t, err)
	assert.Empty(t, resp.GetRecentMessages())
	assert.Equal(t, "first message ever", resp.GetContent())
}
