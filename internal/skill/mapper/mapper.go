// internal/skill/mapper/mapper.go
package mapper

import (
	"portfolio-ai/internal/skill/entity"
	techMapper "portfolio-ai/internal/technology/mapper"
	pb "portfolio-ai/proto/skill"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto maps a domain Skill entity to a protobuf Skill message.
func ToProto(e *entity.Skill) *pb.Skill {
	if e == nil {
		return nil
	}
	s := &pb.Skill{
		Id:           e.ID,
		DisplayOrder: e.DisplayOrder,
		TechnologyId: e.TechnologyID,
		Level:        e.Level,
		Years:        e.Years,
		Favorite:     e.Favorite,
		CreatedAt:    timestamppb.New(e.CreatedAt),
	}
	if e.Technology.ID != "" {
		s.Technology = techMapper.ToProto(&e.Technology)
	}
	return s
}

// ToEntityFromCreate maps a CreateSkillRequest to a domain Skill entity.
func ToEntityFromCreate(r *pb.CreateSkillRequest) *entity.Skill {
	if r == nil {
		return nil
	}
	return &entity.Skill{
		DisplayOrder: r.DisplayOrder,
		TechnologyID: r.TechnologyId,
		Level:        r.Level,
		Years:        r.Years,
		Favorite:     r.Favorite,
	}
}

// ToEntityFromUpdate maps an UpdateSkillRequest to a domain Skill entity.
func ToEntityFromUpdate(r *pb.UpdateSkillRequest) *entity.Skill {
	if r == nil {
		return nil
	}
	return &entity.Skill{
		ID:           r.Id,
		DisplayOrder: r.DisplayOrder,
		TechnologyID: r.TechnologyId,
		Level:        r.Level,
		Years:        r.Years,
		Favorite:     r.Favorite,
	}
}
