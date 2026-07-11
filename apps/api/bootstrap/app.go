// apps/api/bootstrap/app.go
package bootstrap

import (
	"fmt"
	"net"

	"portfolio-ai/internal/auth/jwt"
	authgrpc "portfolio-ai/internal/auth/grpc"
	aimodelrepo "portfolio-ai/internal/aimodel/repository"
	aimodelsvc "portfolio-ai/internal/aimodel/service"
	aimodelgrpc "portfolio-ai/internal/aimodel/grpc"
	chatrepo "portfolio-ai/internal/chat/repository"
	chatsvc "portfolio-ai/internal/chat/service"
	chatgrpc "portfolio-ai/internal/chat/grpc"
	profilerepo "portfolio-ai/internal/profile/repository"
	profilesvc "portfolio-ai/internal/profile/service"
	profilegrpc "portfolio-ai/internal/profile/grpc"
	projectrepo "portfolio-ai/internal/project/repository"
	projectsvc "portfolio-ai/internal/project/service"
	projectgrpc "portfolio-ai/internal/project/grpc"
	promptrepo "portfolio-ai/internal/prompt/repository"
	promptsvc "portfolio-ai/internal/prompt/service"
	promptgrpc "portfolio-ai/internal/prompt/grpc"
	visitorrepo "portfolio-ai/internal/visitor/repository"
	visitorsvc "portfolio-ai/internal/visitor/service"
	visitorgrpc "portfolio-ai/internal/visitor/grpc"
	"portfolio-ai/pkg/config"
	"portfolio-ai/pkg/logger"
	aimodelpb "portfolio-ai/proto/aimodel"
	pb "portfolio-ai/proto/auth"
	chatpb "portfolio-ai/proto/chat"
	profilepb "portfolio-ai/proto/profile"
	projectpb "portfolio-ai/proto/project"
	promptpb "portfolio-ai/proto/prompt"
	visitorpb "portfolio-ai/proto/visitor"

	"google.golang.org/grpc"
	gormdb "gorm.io/gorm"
)

type App struct {
	Config     *config.Config
	DB         *gormdb.DB
	GRPCServer *grpc.Server
}

// NewApp initializes configuration, logger, database, and gRPC server.
func NewApp() (*App, error) {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Initialize structured logger
	logger.Init(cfg.App.Env)
	logger.Info("Starting application", "app_name", cfg.App.Name, "env", cfg.App.Env)

	// 3. Connect to database
	db, err := NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}
	logger.Info("Database connection established", "host", cfg.DB.Host, "database", cfg.DB.Name)

	// 4. Initialize JWT manager
	jwtManager := jwt.NewManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)
	logger.Info("JWT manager initialized", "expiry", cfg.Auth.JWTExpiry.String())

	// 5. Create gRPC server (with auth interceptor)
	grpcServer := NewGRPCServer(jwtManager)
	logger.Info("gRPC server initialized")

	// 6. Register Auth service (login without DB)
	authHandler := authgrpc.NewHandler(jwtManager, cfg.Auth.AdminUsername, cfg.Auth.AdminPassword, cfg.Auth.JWTExpiry)
	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	logger.Info("Auth service registered")

	// 7. Initialize Profile module
	profileRepo := profilerepo.NewPostgresRepository(db)
	profileService := profilesvc.NewService(profileRepo)
	profileHandler := profilegrpc.NewHandler(profileService)
	profilepb.RegisterProfileServiceServer(grpcServer, profileHandler)
	logger.Info("Profile service registered")

	// 8. Initialize Prompt module
	promptRepo := promptrepo.NewPostgresRepository(db)
	promptService := promptsvc.NewService(promptRepo)
	promptHandler := promptgrpc.NewHandler(promptService)
	promptpb.RegisterPromptServiceServer(grpcServer, promptHandler)
	logger.Info("Prompt service registered")

	// 9. Initialize Visitor module
	visitorRepo := visitorrepo.NewPostgresRepository(db)
	visitorService := visitorsvc.NewService(visitorRepo)
	visitorHandler := visitorgrpc.NewHandler(visitorService)
	visitorpb.RegisterVisitorServiceServer(grpcServer, visitorHandler)
	logger.Info("Visitor service registered")

	// 10. Initialize Chat module
	chatRepo := chatrepo.NewPostgresRepository(db)
	chatService := chatsvc.NewService(chatRepo)
	chatHandler := chatgrpc.NewHandler(chatService, visitorService)
	chatpb.RegisterChatServiceServer(grpcServer, chatHandler)
	logger.Info("Chat service registered")

	// 11. Initialize AIModel module
	aimodelRepo := aimodelrepo.NewPostgresRepository(db)
	aimodelService := aimodelsvc.NewService(aimodelRepo)
	aimodelHandler := aimodelgrpc.NewHandler(aimodelService)
	aimodelpb.RegisterAIModelServiceServer(grpcServer, aimodelHandler)
	logger.Info("AIModel service registered")

	// 12. Initialize Project module
	projectRepo := projectrepo.NewPostgresRepository(db)
	projectService := projectsvc.NewService(projectRepo)
	projectHandler := projectgrpc.NewHandler(projectService)
	projectpb.RegisterProjectServiceServer(grpcServer, projectHandler)
	logger.Info("Project service registered")

	return &App{
		Config:     cfg,
		DB:         db,
		GRPCServer: grpcServer,
	}, nil
}

// Run starts the gRPC server on the configured port.
func (a *App) Run() error {
	addr := ":" + a.Config.App.Port
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	logger.Info("gRPC server listening", "address", addr)
	return a.GRPCServer.Serve(lis)
}
