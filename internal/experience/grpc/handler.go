// internal/experience/grpc/handler.go
package grpc

import (
	"context"
	"errors"
	"portfolio-ai/internal/experience/mapper"
	"portfolio-ai/internal/experience/service"
	pb "portfolio-ai/proto/experience"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Handler struct {
	pb.UnimplementedExperienceServiceServer
	svc service.Service
}

// NewHandler creates a new gRPC handler for the Experience service.
func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListExperiences(ctx context.Context, req *pb.ListExperiencesRequest) (*pb.ListExperiencesResponse, error) {
	page := int(req.GetPage())
	limit := int(req.GetLimit())
	if page <= 0 {
		page = 1
	}

	experiences, total, err := h.svc.List(ctx, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list experiences: %v", err)
	}

	pbExperiences := make([]*pb.Experience, 0, len(experiences))
	for _, e := range experiences {
		pbExperiences = append(pbExperiences, mapper.ToProto(e))
	}

	return &pb.ListExperiencesResponse{
		Experiences: pbExperiences,
		Total:       int32(total),
	}, nil
}

func (h *Handler) GetExperience(ctx context.Context, req *pb.GetExperienceRequest) (*pb.GetExperienceResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	e, err := h.svc.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "experience not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get experience: %v", err)
	}

	return &pb.GetExperienceResponse{
		Experience: mapper.ToProto(e),
	}, nil
}

func (h *Handler) CreateExperience(ctx context.Context, req *pb.CreateExperienceRequest) (*pb.CreateExperienceResponse, error) {
	if req.GetCompany() == "" {
		return nil, status.Error(codes.InvalidArgument, "company is required")
	}
	if req.GetPosition() == "" {
		return nil, status.Error(codes.InvalidArgument, "position is required")
	}

	e := mapper.ToEntityFromCreate(req)
	if err := h.svc.Create(ctx, e); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create experience: %v", err)
	}

	return &pb.CreateExperienceResponse{
		Experience: mapper.ToProto(e),
	}, nil
}

func (h *Handler) UpdateExperience(ctx context.Context, req *pb.UpdateExperienceRequest) (*pb.UpdateExperienceResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetCompany() == "" {
		return nil, status.Error(codes.InvalidArgument, "company is required")
	}
	if req.GetPosition() == "" {
		return nil, status.Error(codes.InvalidArgument, "position is required")
	}

	e := mapper.ToEntityFromUpdate(req)
	if err := h.svc.Update(ctx, e); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "experience not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update experience: %v", err)
	}

	return &pb.UpdateExperienceResponse{
		Experience: mapper.ToProto(e),
	}, nil
}

func (h *Handler) DeleteExperience(ctx context.Context, req *pb.DeleteExperienceRequest) (*pb.DeleteExperienceResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "experience not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete experience: %v", err)
	}

	return &pb.DeleteExperienceResponse{
		Success: true,
	}, nil
}
