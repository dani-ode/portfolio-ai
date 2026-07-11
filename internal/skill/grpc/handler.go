// internal/skill/grpc/handler.go
package grpc

import (
	"context"
	"errors"
	"portfolio-ai/internal/skill/mapper"
	"portfolio-ai/internal/skill/service"
	pb "portfolio-ai/proto/skill"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Handler struct {
	pb.UnimplementedSkillServiceServer
	svc service.Service
}

// NewHandler creates a new gRPC handler for the Skill service.
func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListSkills(ctx context.Context, req *pb.ListSkillsRequest) (*pb.ListSkillsResponse, error) {
	page := int(req.GetPage())
	limit := int(req.GetLimit())
	if page <= 0 {
		page = 1
	}

	skills, total, err := h.svc.List(ctx, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list skills: %v", err)
	}

	pbSkills := make([]*pb.Skill, 0, len(skills))
	for _, e := range skills {
		pbSkills = append(pbSkills, mapper.ToProto(e))
	}

	return &pb.ListSkillsResponse{
		Skills: pbSkills,
		Total:  int32(total),
	}, nil
}

func (h *Handler) GetSkill(ctx context.Context, req *pb.GetSkillRequest) (*pb.GetSkillResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	e, err := h.svc.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "skill not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get skill: %v", err)
	}

	return &pb.GetSkillResponse{
		Skill: mapper.ToProto(e),
	}, nil
}

func (h *Handler) CreateSkill(ctx context.Context, req *pb.CreateSkillRequest) (*pb.CreateSkillResponse, error) {
	if req.GetTechnologyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "technology_id is required")
	}

	e := mapper.ToEntityFromCreate(req)
	if err := h.svc.Create(ctx, e); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create skill: %v", err)
	}

	return &pb.CreateSkillResponse{
		Skill: mapper.ToProto(e),
	}, nil
}

func (h *Handler) UpdateSkill(ctx context.Context, req *pb.UpdateSkillRequest) (*pb.UpdateSkillResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetTechnologyId() == "" {
		return nil, status.Error(codes.InvalidArgument, "technology_id is required")
	}

	e := mapper.ToEntityFromUpdate(req)
	if err := h.svc.Update(ctx, e); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "skill not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update skill: %v", err)
	}

	return &pb.UpdateSkillResponse{
		Skill: mapper.ToProto(e),
	}, nil
}

func (h *Handler) DeleteSkill(ctx context.Context, req *pb.DeleteSkillRequest) (*pb.DeleteSkillResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "skill not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete skill: %v", err)
	}

	return &pb.DeleteSkillResponse{
		Success: true,
	}, nil
}
