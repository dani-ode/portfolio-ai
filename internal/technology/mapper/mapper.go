// internal/technology/mapper/mapper.go
package mapper

import (
	"portfolio-ai/internal/technology/entity"
	pb "portfolio-ai/proto/technology"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto maps a domain Technology entity to a protobuf Technology message.
func ToProto(e *entity.Technology) *pb.Technology {
	if e == nil {
		return nil
	}
	return &pb.Technology{
		Id:          e.ID,
		Name:        e.Name,
		Category:    e.Category,
		Icon:        e.Icon,
		Color:       e.Color,
		OfficialUrl: e.OfficialURL,
		Logo:        e.Logo,
		CreatedAt:   timestamppb.New(e.CreatedAt),
	}
}

// ToEntityFromCreate maps a CreateTechnologyRequest to a domain Technology entity.
func ToEntityFromCreate(r *pb.CreateTechnologyRequest) *entity.Technology {
	if r == nil {
		return nil
	}
	return &entity.Technology{
		Name:        r.Name,
		Category:    r.Category,
		Icon:        r.Icon,
		Color:       r.Color,
		OfficialURL: r.OfficialUrl,
		Logo:        r.Logo,
	}
}

// ToEntityFromUpdate maps an UpdateTechnologyRequest to a domain Technology entity.
func ToEntityFromUpdate(r *pb.UpdateTechnologyRequest) *entity.Technology {
	if r == nil {
		return nil
	}
	return &entity.Technology{
		ID:          r.Id,
		Name:        r.Name,
		Category:    r.Category,
		Icon:        r.Icon,
		Color:       r.Color,
		OfficialURL: r.OfficialUrl,
		Logo:        r.Logo,
	}
}
