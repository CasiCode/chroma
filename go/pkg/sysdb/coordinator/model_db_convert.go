package coordinator

import (
	"github.com/chroma-core/chroma/go/pkg/sysdb/coordinator/model"
	"github.com/chroma-core/chroma/go/pkg/sysdb/metastore/db/dbmodel"
	"github.com/chroma-core/chroma/go/pkg/types"
	"github.com/pingcap/log"
	"go.uber.org/zap"
)

func convertCollectionToModel(collectionAndMetadataList []*dbmodel.CollectionAndMetadata) []*model.Collection {
	if collectionAndMetadataList == nil {
		return nil
	}
	collections := make([]*model.Collection, 0, len(collectionAndMetadataList))
	for _, collectionAndMetadata := range collectionAndMetadataList {
		var rootCollectionID *types.UniqueID
		if collectionAndMetadata.Collection.RootCollectionId != nil {
			if id, err := types.Parse(*collectionAndMetadata.Collection.RootCollectionId); err == nil {
				rootCollectionID = &id
			}
		}
		collection := &model.Collection{
			ID:                         types.MustParse(collectionAndMetadata.Collection.ID),
			Name:                       *collectionAndMetadata.Collection.Name,
			ConfigurationJsonStr:       *collectionAndMetadata.Collection.ConfigurationJsonStr,
			Dimension:                  collectionAndMetadata.Collection.Dimension,
			TenantID:                   collectionAndMetadata.TenantID,
			DatabaseName:               collectionAndMetadata.DatabaseName,
			Ts:                         collectionAndMetadata.Collection.Ts,
			LogPosition:                collectionAndMetadata.Collection.LogPosition,
			Version:                    collectionAndMetadata.Collection.Version,
			TotalRecordsPostCompaction: collectionAndMetadata.Collection.TotalRecordsPostCompaction,
			SizeBytesPostCompaction:    collectionAndMetadata.Collection.SizeBytesPostCompaction,
			LastCompactionTimeSecs:     collectionAndMetadata.Collection.LastCompactionTimeSecs,
			RootCollectionID:           rootCollectionID,
			LineageFileName:            collectionAndMetadata.Collection.LineageFileName,
			IsDeleted:                  collectionAndMetadata.Collection.IsDeleted,
			VersionFileName:            collectionAndMetadata.Collection.VersionFileName,
			CreatedAt:                  collectionAndMetadata.Collection.CreatedAt,
			UpdatedAt:                  collectionAndMetadata.Collection.UpdatedAt.Unix(),
			DatabaseId:                 types.MustParse(collectionAndMetadata.Collection.DatabaseID),
		}
		collection.Metadata = convertCollectionMetadataToModel(collectionAndMetadata.CollectionMetadata)
		collections = append(collections, collection)
	}
	log.Debug("collection to model", zap.Any("collections", collections))
	return collections
}

func convertCollectionToGcToModel(collectionToGc []*dbmodel.CollectionToGc) []*model.CollectionToGc {
	if collectionToGc == nil {
		return nil
	}
	collections := make([]*model.CollectionToGc, 0, len(collectionToGc))
	for _, collectionInfo := range collectionToGc {
		collection := model.CollectionToGc{
			ID:              types.MustParse(collectionInfo.ID),
			Name:            collectionInfo.Name,
			VersionFilePath: collectionInfo.VersionFileName,
			TenantID:        collectionInfo.TenantID,
			LineageFilePath: collectionInfo.LineageFileName,
		}
		collections = append(collections, &collection)
	}
	return collections
}

func convertCollectionMetadataToModel(collectionMetadataList []*dbmodel.CollectionMetadata) *model.CollectionMetadata[model.CollectionMetadataValueType] {
	metadata := model.NewCollectionMetadata[model.CollectionMetadataValueType]()
	if collectionMetadataList == nil {
		log.Debug("collection metadata to model", zap.Any("collectionMetadata", nil))
		return nil
	} else {
		for _, collectionMetadata := range collectionMetadataList {
			if collectionMetadata.Key != nil {
				switch {
				case collectionMetadata.BoolValue != nil:
					metadata.Add(*collectionMetadata.Key, &model.CollectionMetadataValueBoolType{Value: *collectionMetadata.BoolValue})
				case collectionMetadata.StrValue != nil:
					metadata.Add(*collectionMetadata.Key, &model.CollectionMetadataValueStringType{Value: *collectionMetadata.StrValue})
				case collectionMetadata.IntValue != nil:
					metadata.Add(*collectionMetadata.Key, &model.CollectionMetadataValueInt64Type{Value: *collectionMetadata.IntValue})
				case collectionMetadata.FloatValue != nil:
					metadata.Add(*collectionMetadata.Key, &model.CollectionMetadataValueFloat64Type{Value: *collectionMetadata.FloatValue})
				default:
				}
			}
		}
		if metadata.Empty() {
			metadata = nil
		}
		log.Debug("collection metadata to model", zap.Any("collectionMetadata", metadata))
		return metadata
	}

}

func convertCollectionMetadataToDB(collectionID string, metadata *model.CollectionMetadata[model.CollectionMetadataValueType]) []*dbmodel.CollectionMetadata {
	if metadata == nil {
		log.Debug("collection metadata to db", zap.Any("collectionMetadata", nil))
		return nil
	}
	dbCollectionMetadataList := make([]*dbmodel.CollectionMetadata, 0, len(metadata.Metadata))
	for key, value := range metadata.Metadata {
		keyCopy := key
		dbCollectionMetadata := &dbmodel.CollectionMetadata{
			CollectionID: collectionID,
			Key:          &keyCopy,
		}
		switch v := (value).(type) {
		case *model.CollectionMetadataValueBoolType:
			dbCollectionMetadata.BoolValue = &v.Value
		case *model.CollectionMetadataValueStringType:
			dbCollectionMetadata.StrValue = &v.Value
		case *model.CollectionMetadataValueInt64Type:
			dbCollectionMetadata.IntValue = &v.Value
		case *model.CollectionMetadataValueFloat64Type:
			dbCollectionMetadata.FloatValue = &v.Value
		default:
			log.Error("unknown collection metadata type", zap.Any("value", v))
		}
		dbCollectionMetadataList = append(dbCollectionMetadataList, dbCollectionMetadata)
	}
	log.Debug("collection metadata to db", zap.Any("collectionMetadata", dbCollectionMetadataList))
	return dbCollectionMetadataList
}

func convertSegmentToModel(segmentAndMetadataList []*dbmodel.SegmentAndMetadata) []*model.Segment {
	if segmentAndMetadataList == nil {
		return nil
	}
	segments := make([]*model.Segment, 0, len(segmentAndMetadataList))
	for _, segmentAndMetadata := range segmentAndMetadataList {
		segment := &model.Segment{
			ID:    types.MustParse(segmentAndMetadata.Segment.ID),
			Type:  segmentAndMetadata.Segment.Type,
			Scope: segmentAndMetadata.Segment.Scope,
			Ts:    segmentAndMetadata.Segment.Ts,
		}
		if segmentAndMetadata.Segment.CollectionID != nil {
			segment.CollectionID = types.MustParse(*segmentAndMetadata.Segment.CollectionID)
		} else {
			segment.CollectionID = types.NilUniqueID()
		}

		segment.Metadata = convertSegmentMetadataToModel(segmentAndMetadata.SegmentMetadata)
		segments = append(segments, segment)
	}
	log.Debug("segment to model", zap.Any("segments", segments))
	return segments
}

func convertSegmentMetadataToModel(segmentMetadataList []*dbmodel.SegmentMetadata) *model.SegmentMetadata[model.SegmentMetadataValueType] {
	if segmentMetadataList == nil {
		return nil
	} else {
		metadata := model.NewSegmentMetadata[model.SegmentMetadataValueType]()
		for _, segmentMetadata := range segmentMetadataList {
			if segmentMetadata.Key != nil {
				switch {
				case segmentMetadata.BoolValue != nil:
					metadata.Set(*segmentMetadata.Key, &model.SegmentMetadataValueBoolType{Value: *segmentMetadata.BoolValue})
				case segmentMetadata.StrValue != nil:
					metadata.Set(*segmentMetadata.Key, &model.SegmentMetadataValueStringType{Value: *segmentMetadata.StrValue})
				case segmentMetadata.IntValue != nil:
					metadata.Set(*segmentMetadata.Key, &model.SegmentMetadataValueInt64Type{Value: *segmentMetadata.IntValue})
				case segmentMetadata.FloatValue != nil:
					metadata.Set(*segmentMetadata.Key, &model.SegmentMetadataValueFloat64Type{Value: *segmentMetadata.FloatValue})
				default:
				}
			}
		}
		if metadata.Empty() {
			metadata = nil
		}
		log.Debug("segment metadata to model", zap.Any("segmentMetadata", nil))
		return metadata
	}
}

func convertSegmentMetadataToDB(segmentID string, metadata *model.SegmentMetadata[model.SegmentMetadataValueType]) []*dbmodel.SegmentMetadata {
	if metadata == nil {
		log.Debug("segment metadata db", zap.Any("segmentMetadata", nil))
		return nil
	}
	dbSegmentMetadataList := make([]*dbmodel.SegmentMetadata, 0, len(metadata.Metadata))
	for key, value := range metadata.Metadata {
		keyCopy := key
		dbSegmentMetadata := &dbmodel.SegmentMetadata{
			SegmentID: segmentID,
			Key:       &keyCopy,
		}
		switch v := (value).(type) {
		case *model.SegmentMetadataValueBoolType:
			dbSegmentMetadata.BoolValue = &v.Value
		case *model.SegmentMetadataValueStringType:
			dbSegmentMetadata.StrValue = &v.Value
		case *model.SegmentMetadataValueInt64Type:
			dbSegmentMetadata.IntValue = &v.Value
		case *model.SegmentMetadataValueFloat64Type:
			dbSegmentMetadata.FloatValue = &v.Value
		default:
			log.Error("unknown segment metadata type", zap.Any("value", v))
		}
		dbSegmentMetadataList = append(dbSegmentMetadataList, dbSegmentMetadata)
	}
	log.Debug("segment metadata db", zap.Any("segmentMetadata", dbSegmentMetadataList))
	return dbSegmentMetadataList
}

func convertDatabaseToModel(dbDatabase *dbmodel.Database) *model.Database {
	return &model.Database{
		ID:     dbDatabase.ID,
		Name:   dbDatabase.Name,
		Tenant: dbDatabase.TenantID,
	}
}

func convertTenantToModel(dbTenant *dbmodel.Tenant) *model.Tenant {
	var resourceName *string
	if dbTenant.ResourceName != nil {
		resourceName = dbTenant.ResourceName
	}
	return &model.Tenant{
		Name:         dbTenant.ID,
		ResourceName: resourceName,
	}
}
