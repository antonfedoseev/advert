package upload

import (
	"crypto/sha256"
	"fmt"
	"github.com/go-logr/logr"
	"internal/advert"
	"internal/env"
	"internal/global"
	"net/http"
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

	advertJson := r.Header.Get("advert")

	a := &advert.Advert{}
	err := a.Load(advertJson)
	if err != nil {
		msg := "Bad request, bad advert body. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	expectedHash := r.Header.Get("hash")
	calculatedHash := fmt.Sprintf("%s", sha256.Sum256([]byte(advertJson)))

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

	err = r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		msg := "Bad request, cant parse files. Error: " + err.Error()
		http.Error(w, msg, http.StatusBadRequest)
	}

	ctx := r.Context()
	var env = env.NewEnvironment(s.hub)
	defer env.Close()

	err = advert.CreateAdvert(ctx, env, a, r.MultipartForm.File)
	if err != nil {
		msg := "Can't save files. Error: " + err.Error()
		http.Error(w, msg, http.StatusInternalServerError)
	}

	data, err := a.Save()
	if err != nil {
		msg := "Can't parse response. Error: " + err.Error()
		http.Error(w, msg, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func isValidToken(token string) bool {
	/*tokenBytes := []byte(token)
	expectedBytes := []byte(constant.INFO_GATEWAY_TOKEN)
	return subtle.ConstantTimeCompare(tokenBytes, expectedBytes) == 1*/
	return true
}
