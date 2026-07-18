package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func (c *Client) InitCollection(ctx context.Context, knowledgeCollection, visitorCollection string, dim int, metricType string) error {
	if err := c.ensureKnowledgeCollection(ctx, knowledgeCollection, dim, metricType); err != nil {
		return err
	}
	if err := c.ensureVisitorMemoryCollection(ctx, visitorCollection, dim, metricType); err != nil {
		return err
	}
	return nil
}

func getMetricType(metricType string) entity.MetricType {
	switch metricType {
	case "IP":
		return entity.IP
	case "L2":
		return entity.L2
	default:
		return entity.COSINE
	}
}

func (c *Client) ensureKnowledgeCollection(ctx context.Context, collectionName string, dim int, metricType string) error {
	has, err := c.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}
	if has {
		if err := c.LoadCollection(ctx, collectionName, false); err != nil {
			return fmt.Errorf("failed to load collection: %w", err)
		}
		return nil
	}

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "Knowledge chunks for dan semantic search",
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       "chunk_id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				TypeParams: map[string]string{"max_length": "26"},
			},
			{
				Name:       "document_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "26"},
			},
			{
				Name:       "source_type",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "50"},
			},
			{
				Name:       "source_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "26"},
			},
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
		},
	}

	if err := c.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	mType := getMetricType(metricType)
	idx, err := entity.NewIndexAUTOINDEX(mType)
	if err != nil {
		return fmt.Errorf("failed to create index definition: %w", err)
	}

	if err := c.CreateIndex(ctx, collectionName, "embedding", idx, false); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	if err := c.LoadCollection(ctx, collectionName, false); err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	return nil
}

func (c *Client) ensureVisitorMemoryCollection(ctx context.Context, collectionName string, dim int, metricType string) error {
	has, err := c.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check visitor memory collection: %w", err)
	}
	if has {
		if err := c.LoadCollection(ctx, collectionName, false); err != nil {
			return fmt.Errorf("failed to load visitor memory collection: %w", err)
		}
		return nil
	}

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "Visitor memories for personalized retrieval",
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				TypeParams: map[string]string{"max_length": "26"},
			},
			{
				Name:       "visitor_id",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{"max_length": "26"},
			},
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dim),
				},
			},
		},
	}

	if err := c.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		return fmt.Errorf("failed to create visitor memory collection: %w", err)
	}

	mType := getMetricType(metricType)
	idx, err := entity.NewIndexAUTOINDEX(mType)
	if err != nil {
		return fmt.Errorf("failed to create visitor memory index definition: %w", err)
	}

	if err := c.CreateIndex(ctx, collectionName, "embedding", idx, false); err != nil {
		return fmt.Errorf("failed to create visitor memory index: %w", err)
	}

	if err := c.LoadCollection(ctx, collectionName, false); err != nil {
		return fmt.Errorf("failed to load visitor memory collection: %w", err)
	}

	return nil
}

// SetAlias drops any existing alias with the given name and recreates it to point to the specified collection.
func (c *Client) SetAlias(ctx context.Context, collectionName string, aliasName string) error {
	// Drop the alias first (ignore error if it doesn't exist)
	_ = c.DropAlias(ctx, aliasName)

	if err := c.CreateAlias(ctx, collectionName, aliasName); err != nil {
		return fmt.Errorf("failed to create alias %s for collection %s: %w", aliasName, collectionName, err)
	}
	return nil
}

