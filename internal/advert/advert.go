package advert

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"internal/env"
	"internal/geo"
	"mime/multipart"
	"net/url"
	"pkg/db"
	"time"
)

type Status int

const (
	StatusUnknown Status = iota
	StatusCreated
	StatusPrepared
	StatusActive
	StatusRejected
)

type ProductState int

const (
	ProductStateUndefined ProductState = iota
	ProductStateUsed
	ProductStateAsNew
	ProductStateNew
)

type SchemaProductDetails struct {
	advertId     uint32
	state        byte
	price        uint32
	category     byte
	subCategory1 byte
	subCategory2 byte
	subCategory3 byte
	geolocation  geo.Point
	country      uint16
	area         uint16
	city         uint32
	district     byte
}

type SchemaAdvert struct {
	id          uint32
	ownerId     uint32
	title       string
	description string
	сTime       uint32
	sTime       uint32
	fTime       uint32
	state       byte
}

type Advert struct {
	Id             uint32
	OwnerId        uint32
	Title          string
	Description    string
	CTime          uint32
	STime          uint32
	FTime          uint32
	State          Status
	ProductDetails *ProductDetails
	Photos         []*Photo
}

func (a *Advert) Load(value string) error {
	return json.Unmarshal([]byte(value), a)
}

func (a *Advert) Save() ([]byte, error) {
	return json.Marshal(a)
}

type ProductDetails struct {
	State        byte
	Price        uint32
	Category     byte
	SubCategory1 byte
	SubCategory2 byte
	SubCategory3 byte
	Geolocation  geo.Point
	Country      uint16
	Area         uint16
	City         uint32
	District     byte
}

func CreateAdvert(ctx context.Context, env *env.Environment, advert *Advert,
	multiFiles map[string][]*multipart.FileHeader) error {

	photoNames, err := storeUserPhotos(ctx, env, uint(advert.OwnerId), uint(advert.Id), multiFiles)
	if err != nil {
		return err
	}

	if len(photoNames) == 0 {
		return errors.New("No photos have been got")
	}

	dbConn, err := env.ShardDb(advert.OwnerId)
	if err != nil {
		return err
	}

	advert.CTime = uint32(time.Now().Unix())
	advert.State = StatusCreated

	schemaAdvert := convertAdvertBusinessToDb(advert)
	schemaProductDetails := convertProductDetailsBusinessToDb(advert.Id, advert.ProductDetails)
	schemaProductPhotos := buildProductPhotosByNames(env.Settings.StaticStorage.Url, advert.Id, photoNames)

	err = dbConn.Transaction(func(conn *db.Conn) error {
		{
			err := createAdvert(conn, schemaAdvert)
			if err != nil {
				return err
			}
		}

		{
			err := createProductDetails(conn, schemaProductDetails)
			if err != nil {
				return err
			}
		}

		{
			err := createProductPhotos(conn, schemaProductPhotos)
			if err != nil {
				return err
			}
		}

		{
			err := sendProcessPhotosRequestToMb(ctx, env, advert.OwnerId, advert.Id, schemaProductPhotos)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	advert.Photos = convertProductPhotosDbToBusiness(schemaProductPhotos)

	return err
}

func sendProcessPhotosRequestToMb(ctx context.Context, env *env.Environment,
	ownerId uint32, advertId uint32, photos []*SchemaPhoto) error {

	req := &processPhotoRequest{
		advertId: advertId,
		ownerId:  ownerId,
		photos:   getProcessPhotosInfo(photos),
	}

	producer := env.MbProducer()
	err := producer.SendMessage(ctx, "advert_process_photo_request",
		fmt.Sprintf("%d_%d", ownerId, advertId), req)

	return err
}

func buildProductPhotosByNames(hostUrl string, advertId uint32, names []string) []*SchemaPhoto {
	photos := make([]*SchemaPhoto, len(names))
	for i := 0; i < len(names); i++ {
		id := i + 1
		url, _ := url.JoinPath(hostUrl, names[i])
		photos[i] = &SchemaPhoto{Id: uint32(id), AdvertId: advertId, Url: url, Order: byte(id)}
	}
	return photos
}

func createProductPhotos(conn *db.Conn, photos []*SchemaPhoto) error {
	builder := conn.InsertInto("product_photo").
		Cols("id", "advert_id", "url", "order")

	for _, photo := range photos {
		builder.Values(photo.Id, photo.AdvertId, photo.Url, photo.Order)
	}

	_, err := builder.Exec()
	return err
}

func createProductDetails(conn *db.Conn, details *SchemaProductDetails) error {
	_, err := conn.InsertInto("product_details").
		Cols("advert_id", "state", "price", "category", "sub_category_1",
			"sub_category_2", "sub_category_3", "geolocation", "country", "area",
			"city", "district").
		Values(details.advertId, details.state, details.price,
			details.category, details.subCategory1, details.subCategory2,
			details.subCategory3, details.geolocation, details.country,
			details.area, details.city, details.district).
		Exec()

	return err
}

func createAdvert(conn *db.Conn, advert *SchemaAdvert) error {
	_, err := conn.InsertInto("advert").
		Cols("id", "owner_id", "title", "description", "сtime", "state").
		Values(advert.id, advert.ownerId, advert.title, advert.description,
			advert.сTime, advert.state).
		Exec()

	return err
}

func updateAdvertState(dbConn *db.Conn, ownerId uint32, advertId uint32, status Status) error {
	ub := dbConn.Update("advert")
	_, err := ub.Set(ub.Assign("state", status)).
		Where(
			ub.Equal("id", advertId),
			ub.Equal("owner_id", ownerId)).
		Exec()

	return err
}

func convertProductPhotosDbToBusiness(schemaPhotos []*SchemaPhoto) []*Photo {
	photos := make([]*Photo, 0, len(schemaPhotos))
	for i := 0; i < len(schemaPhotos); i++ {
		photos = append(photos, convertProductPhotoDbToBusiness(schemaPhotos[i]))
	}
	return photos
}

func convertProductPhotoDbToBusiness(schemaPhoto *SchemaPhoto) *Photo {
	return &Photo{
		schemaPhoto.Id,
		schemaPhoto.Url,
		schemaPhoto.UrlSmall,
		schemaPhoto.UrlMedium,
		schemaPhoto.UrlBig,
		schemaPhoto.Order,
	}
}

func convertAdvertBusinessToDb(advert *Advert) *SchemaAdvert {
	return &SchemaAdvert{
		advert.Id,
		advert.OwnerId,
		advert.Title,
		advert.Description,
		advert.CTime,
		advert.STime,
		advert.FTime,
		byte(advert.State),
	}
}

func convertProductDetailsBusinessToDb(advertId uint32, details *ProductDetails) *SchemaProductDetails {
	return &SchemaProductDetails{
		advertId,
		details.State,
		details.Price,
		details.Category,
		details.SubCategory1,
		details.SubCategory2,
		details.SubCategory3,
		details.Geolocation,
		details.Country,
		details.Area,
		details.City,
		details.District,
	}
}
