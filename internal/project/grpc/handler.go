// internal/project/grpc/handler.go
package grpc

import (
	"context"
	"errors"
	"portfolio-ai/internal/project/mapper"
	"portfolio-ai/internal/project/service"
	pb "portfolio-ai/proto/project"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Handler struct {
	pb.UnimplementedProjectServiceServer
	svc service.Service
}

// NewHandler creates a new gRPC handler for the Project service.
func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListProjects(ctx context.Context, req *pb.ListProjectsRequest) (*pb.ListProjectsResponse, error) {
	page := int(req.GetPage())
	limit := int(req.GetLimit())
	if page <= 0 {
		page = 1
	}

	projects, total, err := h.svc.List(ctx, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list projects: %v", err)
	}

	pbProjects := make([]*pb.Project, 0, len(projects))
	for _, p := range projects {
		pbProjects = append(pbProjects, mapper.ToProto(p))
	}

	return &pb.ListProjectsResponse{
		Projects: pbProjects,
		Total:    int32(total),
	}, nil
}

func (h *Handler) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.GetProjectResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	p, err := h.svc.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get project: %v", err)
	}

	return &pb.GetProjectResponse{
		Project: mapper.ToProto(p),
	}, nil
}

func (h *Handler) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	if req.GetSlug() == "" {
		return nil, status.Error(codes.InvalidArgument, "slug is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	p := mapper.ToEntityFromCreate(req)
	if err := h.svc.Create(ctx, p); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create project: %v", err)
	}

	return &pb.CreateProjectResponse{
		Project: mapper.ToProto(p),
	}, nil
}

func (h *Handler) UpdateProject(ctx context.Context, req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetSlug() == "" {
		return nil, status.Error(codes.InvalidArgument, "slug is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	p := mapper.ToEntityFromUpdate(req)
	if err := h.svc.Update(ctx, p); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update project: %v", err)
	}

	return &pb.UpdateProjectResponse{
		Project: mapper.ToProto(p),
	}, nil
}

func (h *Handler) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete project: %v", err)
	}

	return &pb.DeleteProjectResponse{
		Success: true,
	}, nil
}
