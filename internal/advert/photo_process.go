package advert

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"internal/env"
	"pkg/db"
)

type ProcessPhotoInfo struct {
	id        uint32
	url       string
	urlSmall  string
	urlMedium string
	urlBig    string
}

type processPhotoRequest struct {
	advertId uint32
	ownerId  uint32
	photos   []*ProcessPhotoInfo
}

type ProcessPhotoResponse struct {
	advertId uint32
	ownerId  uint32
	photos   []*ProcessPhotoInfo
}

func (r *ProcessPhotoResponse) Load(value []byte) error {
	return json.Unmarshal(value, r)
}

func ResponsePhotoProcess(env *env.Environment, m *kafka.Message) error {
	response := &ProcessPhotoResponse{}
	err := response.Load(m.Value)
	if err != nil {
		return err
	}

	shardDB, err := env.ShardDb(response.ownerId)
	if err != nil {
		return err
	}

	var urls []string
	sb := shardDB.Select("url")
	_, err = sb.From("product_photo").
		Where(sb.Equal("advert_id", response.advertId)).LoadValues(urls)

	if err != nil {
		return err
	}

	err = shardDB.Transaction(func(dbConn *db.Conn) error {
		//1. Set photo urls in product_photo database
		{
			ub := buildUpdatePhotosQuery(dbConn, response.advertId, response.photos)
			_, err := ub.Exec()
			if err != nil {
				return err
			}
		}

		//2. Change status of advert in advert database
		{
			err := updateAdvertState(dbConn, response.ownerId, response.advertId, StatusPrepared)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	//3. Remove temp photos by url from storage
	err = removeFilesByUrl(env.Settings.StaticStorage.Path, urls)
	if err != nil {
		return err
	}

	return nil
}

func getProcessPhotosInfo(schemaPhotos []*SchemaPhoto) []*ProcessPhotoInfo {
	list := make([]*ProcessPhotoInfo, len(schemaPhotos))
	for i, schema := range schemaPhotos {
		list[i] = &ProcessPhotoInfo{
			id:  schema.Id,
			url: schema.Url,
		}
	}

	return list
}

type SchemaPhoto struct {
	Id        uint32
	AdvertId  uint32
	Url       string
	UrlSmall  string
	UrlMedium string
	UrlBig    string
	Order     byte
}

type Photo struct {
	Id        uint32
	Url       string
	UrlSmall  string
	UrlMedium string
	UrlBig    string
	Order     byte
}

func buildUpdatePhotosQuery(db *db.Conn, advertId uint32, photos []*ProcessPhotoInfo) *db.InsertBuilder {
	builder := db.ReplaceInto("product_photo").
		Cols("id", "advert_id", "url", "url_small", "url_medium", "url_big")

	for _, photo := range photos {
		builder.Values(photo.id, advertId, photo.url, photo.urlSmall, photo.urlMedium, photo.urlBig)
	}

	return builder
}
