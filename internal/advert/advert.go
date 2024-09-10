package advert

import (
	"context"
	"database/sql"
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

var (
	ErrDuplicateAdvert = errors.New("Advert already exist")
)

type SchemaProductDetails struct {
	AdvertId     uint32   `db:"advert_id"`
	State        byte     `db:"state"`
	Price        uint32   `db:"price"`
	Category     byte     `db:"category"`
	SubCategory1 byte     `db:"sub_category_1"`
	SubCategory2 byte     `db:"sub_category_2"`
	SubCategory3 byte     `db:"sub_category_3"`
	Geolocation  db.Point `db:"geolocation"`
	Country      uint16   `db:"country"`
	Area         uint16   `db:"area"`
	City         uint32   `db:"city"`
	District     byte     `db:"district"`
}

type SchemaAdvert struct {
	Id          uint32 `db:"id"`
	OwnerId     uint32 `db:"owner_id"`
	Title       string `db:"title"`
	Description string `db:"description"`
	CTime       uint32 `db:"ctime"`
	STime       uint32 `db:"stime"`
	FTime       uint32 `db:"ftime"`
	State       byte   `db:"state"`
}

type Advert struct {
	Id             uint32          `json:"id"`
	OwnerId        uint32          `json:"owner_id"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	CTime          uint32          `json:"ctime"`
	STime          uint32          `json:"stime"`
	FTime          uint32          `json:"ftime"`
	State          Status          `json:"state"`
	ProductDetails *ProductDetails `json:"product_details"`
	Photos         []*Photo
}

func (a *Advert) Load(value string) error {
	return json.Unmarshal([]byte(value), a)
}

func (a *Advert) Save() ([]byte, error) {
	return json.Marshal(a)
}

type ProductDetails struct {
	State        byte      `json:"state"`
	Price        uint32    `json:"price"`
	Category     byte      `json:"category"`
	SubCategory1 byte      `json:"sub_category_1"`
	SubCategory2 byte      `json:"sub_category_2"`
	SubCategory3 byte      `json:"sub_category_3"`
	Geolocation  geo.Point `json:"geolocation"`
	Country      uint16    `json:"country"`
	Area         uint16    `json:"area"`
	City         uint32    `json:"city"`
	District     byte      `json:"district"`
}

func CreateAdvert(ctx context.Context, env *env.Environment, advert *Advert,
	multiFiles []*multipart.FileHeader) error {

	dbConn, err := env.ShardDb(advert.OwnerId)
	if err != nil {
		return err
	}

	existingAdvert, err := getAdvert(dbConn, advert.Id, advert.OwnerId)
	if err != nil {
		return err
	}

	if existingAdvert != nil {
		return errors.Wrapf(ErrDuplicateAdvert, "advert Id %d, owner Id %d", advert.Id, advert.OwnerId)
	}

	photoNames, err := storeUserPhotos(ctx, env, uint(advert.OwnerId), uint(advert.Id), multiFiles)
	if err != nil {
		return err
	}

	if len(photoNames) == 0 {
		return errors.New("No Photos have been got")
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

func getAdvert(conn *db.Conn, id uint32, ownerId uint32) (*SchemaAdvert, error) {
	var advert SchemaAdvert
	sb := conn.Select("id, owner_id")
	err := sb.From("advert").
		Where(sb.Equal("id", id),
			sb.Equal("owner_id", ownerId),
		).Limit(1).
		LoadStruct(&advert)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &advert, nil
}

func sendProcessPhotosRequestToMb(ctx context.Context, env *env.Environment,
	ownerId uint32, advertId uint32, photos []*SchemaPhoto) error {

	req := &ProcessPhotoRequest{
		AdvertId: advertId,
		OwnerId:  ownerId,
		Photos:   getProcessPhotosInfo(photos),
	}

	producer := env.MbProducer()
	err := producer.SendMessage(
		ctx,
		"advert_process_photo_request",
		fmt.Sprintf("%d_%d", ownerId, advertId),
		req,
	)

	return err
}

func buildProductPhotosByNames(hostUrl string, advertId uint32, names []string) []*SchemaPhoto {
	photos := make([]*SchemaPhoto, len(names))
	for i := 0; i < len(names); i++ {
		id := i + 1
		url, _ := url.JoinPath(hostUrl, names[i])
		photos[i] = &SchemaPhoto{Id: uint32(id), AdvertId: advertId, Url: url, Position: byte(id)}
	}
	return photos
}

func createProductPhotos(conn *db.Conn, photos []*SchemaPhoto) error {
	builder := conn.InsertInto("product_photo").
		Cols("id", "advert_id", "url", "position")

	for _, photo := range photos {
		builder.Values(photo.Id, photo.AdvertId, photo.Url, photo.Position)
	}

	_, err := builder.Exec()
	return err
}

func createProductDetails(conn *db.Conn, details *SchemaProductDetails) error {
	/*_, err := conn.InsertInto("product_details").
	Cols("advert_id", "State", "Price", "Category", "sub_category_1",
		"sub_category_2", "sub_category_3", "Geolocation", "Country", "Area",
		"City", "District").
	Values(details.AdvertId, details.State, details.Price,
		details.Category, details.SubCategory1, details.SubCategory2,
		details.SubCategory3, details.Geolocation, details.Country,
		details.Area, details.City, details.District).
	Exec()*/

	_, err := conn.InsertInto("product_details").
		SQL(
			fmt.Sprintf("(advert_id, state, price, category, sub_category_1, sub_category_2, sub_category_3, geolocation, country, area, city, district) values (%d,%d,%d,%d,%d,%d,%d,%s,%d,%d,%d,%d)",
				details.AdvertId, details.State, details.Price,
				details.Category, details.SubCategory1, details.SubCategory2,
				details.SubCategory3,
				fmt.Sprintf("ST_GeomFromText('POINT(%f %f)')", details.Geolocation.Longitude, details.Geolocation.Latitude),
				details.Country, details.Area, details.City, details.District)).
		Exec()

	return err
}

func createAdvert(conn *db.Conn, advert *SchemaAdvert) error {
	_, err := conn.InsertInto("advert").
		Cols("id", "owner_id", "title", "description", "сtime", "state").
		Values(advert.Id, advert.OwnerId, advert.Title, advert.Description,
			advert.CTime, advert.State).
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
		schemaPhoto.Position,
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
		db.Point{Longitude: details.Geolocation.Longitude, Latitude: details.Geolocation.Latitude},
		details.Country,
		details.Area,
		details.City,
		details.District,
	}
}
