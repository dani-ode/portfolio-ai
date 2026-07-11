// internal/project/mapper/mapper.go
package mapper

import (
	"portfolio-ai/internal/project/entity"
	pb "portfolio-ai/proto/project"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto maps a domain Project entity to a protobuf Project message.
func ToProto(e *entity.Project) *pb.Project {
	if e == nil {
		return nil
	}
	p := &pb.Project{
		Id:           e.ID,
		Slug:         e.Slug,
		Title:        e.Title,
		Summary:      e.Summary,
		Description:  e.Description,
		Architecture: e.Architecture,
		RepositoryUrl: e.RepositoryURL,
		DemoUrl:      e.DemoURL,
		Thumbnail:    e.Thumbnail,
		Featured:     e.Featured,
		Status:       e.Status,
		GithubStars:  e.GithubStars,
		ReadTime:     e.ReadTime,
		CreatedAt:    timestamppb.New(e.CreatedAt),
		UpdatedAt:    timestamppb.New(e.UpdatedAt),
	}
	if e.GithubLastCommit != nil {
		p.GithubLastCommit = timestamppb.New(*e.GithubLastCommit)
	}
	return p
}

// ToEntityFromCreate maps a CreateProjectRequest to a domain Project entity.
func ToEntityFromCreate(r *pb.CreateProjectRequest) *entity.Project {
	if r == nil {
		return nil
	}
	return &entity.Project{
		Slug:          r.Slug,
		Title:         r.Title,
		Summary:       r.Summary,
		Description:   r.Description,
		Architecture:  r.Architecture,
		RepositoryURL: r.RepositoryUrl,
		DemoURL:       r.DemoUrl,
		Thumbnail:     r.Thumbnail,
		Featured:      r.Featured,
		Status:        r.Status,
		GithubStars:   r.GithubStars,
		ReadTime:      r.ReadTime,
	}
}

// ToEntityFromUpdate maps an UpdateProjectRequest to a domain Project entity.
func ToEntityFromUpdate(r *pb.UpdateProjectRequest) *entity.Project {
	if r == nil {
		return nil
	}
	return &entity.Project{
		ID:            r.Id,
		Slug:          r.Slug,
		Title:         r.Title,
		Summary:       r.Summary,
		Description:   r.Description,
		Architecture:  r.Architecture,
		RepositoryURL: r.RepositoryUrl,
		DemoURL:       r.DemoUrl,
		Thumbnail:     r.Thumbnail,
		Featured:      r.Featured,
		Status:        r.Status,
		GithubStars:   r.GithubStars,
		ReadTime:      r.ReadTime,
	}
}
