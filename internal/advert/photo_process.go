package advert

import (
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"internal/env"
	"pkg/db"
)

type ProcessPhotoInfo struct {
	Id        uint32 `json:"id"`
	Url       string `json:"url"`
	UrlSmall  string `json:"url_small"`
	UrlMedium string `json:"url_medium"`
	UrlBig    string `json:"url_big"`
}

type ProcessPhotoRequest struct {
	AdvertId uint32              `json:"advert_id"`
	OwnerId  uint32              `json:"owner_id"`
	Photos   []*ProcessPhotoInfo `json:"photos"`
}

type ProcessPhotoResponse struct {
	AdvertId uint32              `json:"advert_id"`
	OwnerId  uint32              `json:"owner_id"`
	Photos   []*ProcessPhotoInfo `json:"photos"`
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

	shardDB, err := env.ShardDb(response.OwnerId)
	if err != nil {
		return err
	}

	var urls []string
	sb := shardDB.Select("url")
	_, err = sb.From("product_photo").
		Where(sb.Equal("advert_id", response.AdvertId)).LoadValues(&urls)

	if err != nil {
		return err
	}

	err = shardDB.Transaction(func(dbConn *db.Conn) error {
		//1. Set photo urls in product_photo database
		{
			ub := buildUpdatePhotosQuery(dbConn, response.AdvertId, response.Photos)
			_, err := ub.Exec()
			if err != nil {
				return err
			}
		}

		//2. Change status of advert in advert database
		{
			err := updateAdvertState(dbConn, response.OwnerId, response.AdvertId, StatusPrepared)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	//3. Remove temp Photos by Url from storage
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
			Id:  schema.Id,
			Url: schema.Url,
		}
	}

	return list
}

type SchemaPhoto struct {
	Id        uint32 `db:"id"`
	AdvertId  uint32 `db:"advert_id"`
	Url       string `db:"url"`
	UrlSmall  string `db:"url_small"`
	UrlMedium string `db:"url_medium"`
	UrlBig    string `db:"url_big"`
	Position  byte   `db:"position"`
}

type Photo struct {
	Id        uint32 `json:"id"`
	Url       string `json:"url"`
	UrlSmall  string `json:"url_small"`
	UrlMedium string `json:"url_medium"`
	UrlBig    string `json:"url_big"`
	Position  byte   `json:"position"`
}

func buildUpdatePhotosQuery(db *db.Conn, advertId uint32, photos []*ProcessPhotoInfo) *db.InsertBuilder {
	builder := db.ReplaceInto("product_photo").
		Cols("id", "advert_id", "url", "url_small", "url_medium", "url_big")

	for _, photo := range photos {
		builder.Values(photo.Id, advertId, photo.Url, photo.UrlSmall, photo.UrlMedium, photo.UrlBig)
	}

	return builder
}
