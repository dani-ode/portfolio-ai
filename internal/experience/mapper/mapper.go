// internal/experience/mapper/mapper.go
package mapper

import (
	"portfolio-ai/internal/experience/entity"
	pb "portfolio-ai/proto/experience"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto maps a domain Experience entity to a protobuf Experience message.
func ToProto(e *entity.Experience) *pb.Experience {
	if e == nil {
		return nil
	}
	p := &pb.Experience{
		Id:             e.ID,
		Company:        e.Company,
		Position:       e.Position,
		EmploymentType: e.EmploymentType,
		CurrentJob:     e.CurrentJob,
		Location:       e.Location,
		Description:    e.Description,
		DisplayOrder:   e.DisplayOrder,
		CompanyLogo:    e.CompanyLogo,
		Skills:         e.Skills,
		RemoteType:     e.RemoteType,
		CreatedAt:      timestamppb.New(e.CreatedAt),
		UpdatedAt:      timestamppb.New(e.UpdatedAt),
	}

	if e.StartDate != nil {
		p.StartDate = timestamppb.New(*e.StartDate)
	}
	if e.EndDate != nil {
		p.EndDate = timestamppb.New(*e.EndDate)
	}

	return p
}

// ToEntityFromCreate maps a CreateExperienceRequest to a domain Experience entity.
func ToEntityFromCreate(r *pb.CreateExperienceRequest) *entity.Experience {
	if r == nil {
		return nil
	}
	e := &entity.Experience{
		Company:        r.Company,
		Position:       r.Position,
		EmploymentType: r.EmploymentType,
		CurrentJob:     r.CurrentJob,
		Location:       r.Location,
		Description:    r.Description,
		DisplayOrder:   r.DisplayOrder,
		CompanyLogo:    r.CompanyLogo,
		Skills:         r.Skills,
		RemoteType:     r.RemoteType,
	}

	if r.StartDate != nil {
		t := r.StartDate.AsTime()
		e.StartDate = &t
	}
	if r.EndDate != nil {
		t := r.EndDate.AsTime()
		e.EndDate = &t
	}

	return e
}

// ToEntityFromUpdate maps an UpdateExperienceRequest to a domain Experience entity.
func ToEntityFromUpdate(r *pb.UpdateExperienceRequest) *entity.Experience {
	if r == nil {
		return nil
	}
	e := &entity.Experience{
		ID:             r.Id,
		Company:        r.Company,
		Position:       r.Position,
		EmploymentType: r.EmploymentType,
		CurrentJob:     r.CurrentJob,
		Location:       r.Location,
		Description:    r.Description,
		DisplayOrder:   r.DisplayOrder,
		CompanyLogo:    r.CompanyLogo,
		Skills:         r.Skills,
		RemoteType:     r.RemoteType,
	}

	if r.StartDate != nil {
		t := r.StartDate.AsTime()
		e.StartDate = &t
	}
	if r.EndDate != nil {
		t := r.EndDate.AsTime()
		e.EndDate = &t
	}

	return e
}
