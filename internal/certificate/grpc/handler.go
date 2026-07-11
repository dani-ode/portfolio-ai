// internal/certificate/grpc/handler.go
package grpc

import (
	"context"
	"errors"
	"portfolio-ai/internal/certificate/mapper"
	"portfolio-ai/internal/certificate/service"
	pb "portfolio-ai/proto/certificate"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Handler struct {
	pb.UnimplementedCertificateServiceServer
	svc service.Service
}

// NewHandler creates a new gRPC handler for the Certificate service.
func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListCertificates(ctx context.Context, req *pb.ListCertificatesRequest) (*pb.ListCertificatesResponse, error) {
	page := int(req.GetPage())
	limit := int(req.GetLimit())
	if page <= 0 {
		page = 1
	}

	certs, total, err := h.svc.List(ctx, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list certificates: %v", err)
	}

	pbCerts := make([]*pb.Certificate, 0, len(certs))
	for _, e := range certs {
		pbCerts = append(pbCerts, mapper.ToProto(e))
	}

	return &pb.ListCertificatesResponse{
		Certificates: pbCerts,
		Total:        int32(total),
	}, nil
}

func (h *Handler) GetCertificate(ctx context.Context, req *pb.GetCertificateRequest) (*pb.GetCertificateResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	e, err := h.svc.Get(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "certificate not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get certificate: %v", err)
	}

	return &pb.GetCertificateResponse{
		Certificate: mapper.ToProto(e),
	}, nil
}

func (h *Handler) CreateCertificate(ctx context.Context, req *pb.CreateCertificateRequest) (*pb.CreateCertificateResponse, error) {
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetIssuer() == "" {
		return nil, status.Error(codes.InvalidArgument, "issuer is required")
	}

	e := mapper.ToEntityFromCreate(req)
	if err := h.svc.Create(ctx, e); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create certificate: %v", err)
	}

	return &pb.CreateCertificateResponse{
		Certificate: mapper.ToProto(e),
	}, nil
}

func (h *Handler) UpdateCertificate(ctx context.Context, req *pb.UpdateCertificateRequest) (*pb.UpdateCertificateResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetIssuer() == "" {
		return nil, status.Error(codes.InvalidArgument, "issuer is required")
	}

	e := mapper.ToEntityFromUpdate(req)
	if err := h.svc.Update(ctx, e); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "certificate not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update certificate: %v", err)
	}

	return &pb.UpdateCertificateResponse{
		Certificate: mapper.ToProto(e),
	}, nil
}

func (h *Handler) DeleteCertificate(ctx context.Context, req *pb.DeleteCertificateRequest) (*pb.DeleteCertificateResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.svc.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "certificate not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete certificate: %v", err)
	}

	return &pb.DeleteCertificateResponse{
		Success: true,
	}, nil
}
