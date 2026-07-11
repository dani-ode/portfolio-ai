// internal/technology/grpc/handler.go
package grpc

import (
	"context"
	"errors"
	"portfolio-ai/internal/technology/mapper"
	"portfolio-ai/internal/technology/service"
	pb "portfolio-ai/proto/technology"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Handler struct {
	pb.UnimplementedTechnologyServiceServer
	svc service.Service
}

// NewHandler creates a new gRPC handler for the Technology service.
func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListTechnologies(ctx context.Context, req *pb.ListTechnologiesRequest) (*pb.ListTechnologiesResponse, error) {
	page := int(req.GetPage())
	limit := int(req.GetLimit())
	if page <= 0 {
		page = 1
	}

	techs, total, err := h.svc.List(ctx, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list technologies: %v", err)
	}

	pbTechs := make([]*pb.Technology, 0, len(techs))
	for _, e := range techs {
		pbTechs = append(pbTechs, mapper.ToProto(e))
	}

	return &pb.ListTechnologiesResponse{
		Technologies: pbTechs,
		Total:        int32(total),
	}, nil
}

func (h *Handler) GetTechnology(ctx context.Context, req *pb.GetTechnologyRequest) (*pb.GetTechnologyResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	e, err := h.svc.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "technology not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get technology: %v", err)
	}

	return &pb.GetTechnologyResponse{
		Technology: mapper.ToProto(e),
	}, nil
}

func (h *Handler) CreateTechnology(ctx context.Context, req *pb.CreateTechnologyRequest) (*pb.CreateTechnologyResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	e := mapper.ToEntityFromCreate(req)
	if err := h.svc.Create(ctx, e); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create technology: %v", err)
	}

	return &pb.CreateTechnologyResponse{
		Technology: mapper.ToProto(e),
	}, nil
}

func (h *Handler) UpdateTechnology(ctx context.Context, req *pb.UpdateTechnologyRequest) (*pb.UpdateTechnologyResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	e := mapper.ToEntityFromUpdate(req)
	if err := h.svc.Update(ctx, e); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "technology not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update technology: %v", err)
	}

	return &pb.UpdateTechnologyResponse{
		Technology: mapper.ToProto(e),
	}, nil
}

func (h *Handler) DeleteTechnology(ctx context.Context, req *pb.DeleteTechnologyRequest) (*pb.DeleteTechnologyResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "technology not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete technology: %v", err)
	}

	return &pb.DeleteTechnologyResponse{
		Success: true,
	}, nil
}
