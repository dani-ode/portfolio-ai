portfolio-ai/
│
├── .dockerignore
├── .env
├── .env.example
├── .gitignore
├── Makefile
├── README.md
├── apps/
│   ├── api/
│   │   ├── bootstrap/
│   │   │   ├── app.go
│   │   │   ├── database.go
│   │   │   └── grpc.go
│   │   └── main.go
│   ├── worker-embedding/
│   │   ├── bootstrap/
│   │   │   └── worker.go
│   │   └── main.go
│   └── worker-events/
│       ├── bootstrap/
│       │   └── worker.go
│       └── main.go
├── buf.gen.yaml
├── buf.yaml
├── deployments/
│   ├── compose/
│   │   ├── docker-compose.dev.yml
│   │   ├── docker-compose.prod.yml
│   │   └── docker-compose.yml
│   ├── docker/
│   │   ├── api.Dockerfile
│   │   ├── worker-embedding.Dockerfile
│   │   └── worker-events.Dockerfile
│   └── migrations/
│       ├── 000001_create_profiles.down.sql
│       ├── 000001_create_profiles.up.sql
│       ├── 000002_create_remaining_tables.down.sql
│       └── 000002_create_remaining_tables.up.sql
├── docs/
│   └── README.md
├── go.mod
├── go.sum
├── internal/
│   ├── aimodel/
│   │   ├── entity/
│   │   │   └── aimodel.go
│   │   ├── grpc/
│   │   │   └── handler.go
│   │   ├── mapper/
│   │   │   └── mapper.go
│   │   ├── repository/
│   │   │   └── postgres.go
│   │   └── service/
│   │       └── service.go
│   ├── auth/
│   │   ├── grpc/
│   │   │   └── handler.go
│   │   └── jwt/
│   │       └── jwt.go
│   ├── chat/
│   │   ├── entity/
│   │   │   ├── .gitkeep
│   │   │   ├── message.go
│   │   │   └── session.go
│   │   ├── grpc/
│   │   │   ├── .gitkeep
│   │   │   ├── handler.go
│   │   │   └── handler_test.go
│   │   ├── mapper/
│   │   │   ├── .gitkeep
│   │   │   └── mapper.go
│   │   ├── repository/
│   │   │   ├── .gitkeep
│   │   │   └── postgres.go
│   │   └── service/
│   │       ├── .gitkeep
│   │       └── service.go
│   ├── profile/
│   │   ├── entity/
│   │   │   ├── .gitkeep
│   │   │   └── profile.go
│   │   ├── grpc/
│   │   │   ├── .gitkeep
│   │   │   └── handler.go
│   │   ├── mapper/
│   │   │   ├── .gitkeep
│   │   │   └── mapper.go
│   │   ├── repository/
│   │   │   ├── .gitkeep
│   │   │   └── postgres.go
│   │   └── service/
│   │       ├── .gitkeep
│   │       └── service.go
│   ├── prompt/
│   │   ├── entity/
│   │   │   ├── .gitkeep
│   │   │   └── prompt.go
│   │   ├── grpc/
│   │   │   ├── .gitkeep
│   │   │   └── handler.go
│   │   ├── mapper/
│   │   │   ├── .gitkeep
│   │   │   └── mapper.go
│   │   ├── repository/
│   │   │   ├── .gitkeep
│   │   │   └── postgres.go
│   │   └── service/
│   │       ├── .gitkeep
│   │       └── service.go
│   ├── shared/
│   │   ├── constants/
│   │   │   └── constants.go
│   │   ├── errors/
│   │   │   └── errors.go
│   │   ├── interceptor/
│   │   │   ├── .gitkeep
│   │   │   ├── auth.go
│   │   │   ├── logger.go
│   │   │   └── recovery.go
│   │   ├── middleware/
│   │   │   └── .gitkeep
│   │   └── response/
│   │       └── response.go
│   └── visitor/
│       ├── entity/
│       │   ├── .gitkeep
│       │   └── visitor.go
│       ├── grpc/
│       │   ├── .gitkeep
│       │   └── handler.go
│       ├── mapper/
│       │   ├── .gitkeep
│       │   └── mapper.go
│       ├── repository/
│       │   ├── .gitkeep
│       │   └── postgres.go
│       └── service/
│           ├── .gitkeep
│           └── service.go
├── pkg/
│   ├── config/
│   │   └── config.go
│   ├── grpc/
│   │   └── server.go
│   ├── logger/
│   │   └── logger.go
│   ├── postgres/
│   │   └── postgres.go
│   ├── ulid/
│   │   └── ulid.go
│   └── utils/
│       └── .gitkeep
├── proto/
│   ├── README.md
│   ├── aimodel/
│   │   ├── aimodel.pb.go
│   │   ├── aimodel.proto
│   │   ├── aimodel_service.pb.go
│   │   ├── aimodel_service.proto
│   │   └── aimodel_service_grpc.pb.go
│   ├── auth/
│   │   ├── auth.pb.go
│   │   ├── auth.proto
│   │   └── auth_grpc.pb.go
│   ├── chat/
│   │   ├── chat.pb.go
│   │   ├── chat.proto
│   │   ├── chat_service.pb.go
│   │   ├── chat_service.proto
│   │   └── chat_service_grpc.pb.go
│   ├── profile/
│   │   ├── profile.pb.go
│   │   ├── profile.proto
│   │   ├── profile_service.pb.go
│   │   ├── profile_service.proto
│   │   └── profile_service_grpc.pb.go
│   ├── prompt/
│   │   ├── prompt.pb.go
│   │   ├── prompt.proto
│   │   ├── prompt_service.pb.go
│   │   ├── prompt_service.proto
│   │   └── prompt_service_grpc.pb.go
│   └── visitor/
│       ├── visitor.pb.go
│       ├── visitor.proto
│       ├── visitor_service.pb.go
│       ├── visitor_service.proto
│       └── visitor_service_grpc.pb.go
└── scripts/
    └── README.md
