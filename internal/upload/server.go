package upload

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"internal/advert"
	"internal/constant"
	"internal/env"
	"internal/global"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
)

const (
	tokenHeader = "TOKEN"
)

type Server struct {
	hub    global.Hub
	logger logr.Logger
}

func NewServer(globs global.Hub) *Server {
	logger := globs.Logger.WithName("[saveUserFile]")
	return &Server{hub: globs, logger: logger}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get(tokenHeader)
	if len(token) == 0 {
		s.logger.Error(nil, "Can't upload files. Bad request, token is not specified.")
		http.Error(w, "Bad request, token is not specified.", http.StatusBadRequest)
		return
	}

	if !isValidToken(token) {
		http.Error(w, "Bad request, invalid token", http.StatusBadRequest)
		return
	}

	err := r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		msg := "Bad request, cant parse files. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	adverts, ok := r.MultipartForm.Value["advert"]
	if !ok || len(adverts) == 0 {
		msg := "Bad request: no advert body."
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	advertJson := adverts[0]

	a := &advert.Advert{}
	err = a.Load(advertJson)
	if err != nil {
		msg := "Bad request, bad advert body. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	hash, ok := r.MultipartForm.Value["hash"]
	if !ok || len(hash) == 0 {
		msg := "Bad request: no hash."
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	expectedHash := hash[0]
	calculatedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(advertJson)))

	if expectedHash != calculatedHash {
		msg := "Bad request, different hash. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if a.OwnerId <= 0 {
		msg := "Bad request, bad user id. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if a.Id <= 0 {
		msg := "Bad request, bad advert id. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if len(a.Title) == 0 {
		msg := "Bad request, bad title. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if len(a.Description) == 0 {
		msg := "Bad request, bad description. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if a.ProductDetails == nil {
		msg := "Bad request, bad product details body. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if advert.ProductState(a.ProductDetails.State) == advert.ProductStateUndefined {
		msg := "Bad request, bad product state. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if a.ProductDetails.Price < 0 {
		msg := "Bad request, bad product price. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	images, ok := r.MultipartForm.File["images"]
	if !ok || len(images) == 0 {
		msg := "Bad request: no files."
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	err = validateImages(images)
	if err != nil {
		msg := "Bad request, bad image. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var env = env.NewEnvironment(s.hub)
	defer env.Close()

	err = advert.CreateAdvert(ctx, env, a, images)
	if err != nil {
		msg := "Can't save files. Error: " + err.Error()
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	data, err := a.Save()
	if err != nil {
		msg := "Can't parse response. Error: " + err.Error()
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

var supportedExtension = []string{
	"png",
}

func validateImages(images []*multipart.FileHeader) error {
	for _, image := range images {
		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(image.Filename), "."))
		if !slices.Contains(supportedExtension, ext) {
			return errors.Errorf("unsupported image extension \"%s\"", ext)
		}
	}

	return nil
}

func isValidToken(token string) bool {
	tokenBytes := []byte(token)
	expectedBytes := []byte(constant.AdvertGatewayToken)
	return subtle.ConstantTimeCompare(tokenBytes, expectedBytes) == 1
	return true
}
