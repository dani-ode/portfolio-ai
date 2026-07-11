// internal/certificate/mapper/mapper.go
package mapper

import (
	"portfolio-ai/internal/certificate/entity"
	pb "portfolio-ai/proto/certificate"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto maps a domain Certificate entity to a protobuf Certificate message.
func ToProto(e *entity.Certificate) *pb.Certificate {
	if e == nil {
		return nil
	}
	p := &pb.Certificate{
		Id:            e.ID,
		Title:         e.Title,
		Issuer:        e.Issuer,
		CredentialId:  e.CredentialID,
		CredentialUrl: e.CredentialURL,
		Thumbnail:     e.Thumbnail,
		Skills:        e.Skills,
		IssuerLogo:    e.IssuerLogo,
		CreatedAt:     timestamppb.New(e.CreatedAt),
	}

	if e.IssueDate != nil {
		p.IssueDate = timestamppb.New(*e.IssueDate)
	}
	if e.ExpirationDate != nil {
		p.ExpirationDate = timestamppb.New(*e.ExpirationDate)
	}

	return p
}

// ToEntityFromCreate maps a CreateCertificateRequest to a domain Certificate entity.
func ToEntityFromCreate(r *pb.CreateCertificateRequest) *entity.Certificate {
	if r == nil {
		return nil
	}
	e := &entity.Certificate{
		Title:         r.Title,
		Issuer:        r.Issuer,
		CredentialID:  r.CredentialId,
		CredentialURL: r.CredentialUrl,
		Thumbnail:     r.Thumbnail,
		Skills:        r.Skills,
		IssuerLogo:    r.IssuerLogo,
	}

	if r.IssueDate != nil {
		t := r.IssueDate.AsTime()
		e.IssueDate = &t
	}
	if r.ExpirationDate != nil {
		t := r.ExpirationDate.AsTime()
		e.ExpirationDate = &t
	}

	return e
}

// ToEntityFromUpdate maps an UpdateCertificateRequest to a domain Certificate entity.
func ToEntityFromUpdate(r *pb.UpdateCertificateRequest) *entity.Certificate {
	if r == nil {
		return nil
	}
	e := &entity.Certificate{
		ID:            r.Id,
		Title:         r.Title,
		Issuer:        r.Issuer,
		CredentialID:  r.CredentialId,
		CredentialURL: r.CredentialUrl,
		Thumbnail:     r.Thumbnail,
		Skills:        r.Skills,
		IssuerLogo:    r.IssuerLogo,
	}

	if r.IssueDate != nil {
		t := r.IssueDate.AsTime()
		e.IssueDate = &t
	}
	if r.ExpirationDate != nil {
		t := r.ExpirationDate.AsTime()
		e.ExpirationDate = &t
	}

	return e
}
