package advert

import (
	"context"
	"crypto/sha256"
	"fmt"
	"golang.org/x/sync/errgroup"
	"internal/env"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type photoInfo struct {
	Name string
	File *multipart.FileHeader
}

func storeUserPhotos(ctx context.Context, env *env.Environment, userId uint, advertId uint,
	multiFiles map[string][]*multipart.FileHeader) ([]string, error) {

	list := make([]*photoInfo, 0)
	i := 0
	stamp := uint32(time.Now().Unix())

	for _, files := range multiFiles {
		for _, file := range files {
			ext := filepath.Ext(file.Filename)
			userIdHash := sha256.Sum256([]byte(fmt.Sprintf("%d", userId)))
			hash := sha256.Sum256([]byte(fmt.Sprintf("%d_%d_%d", advertId, i, stamp)))
			name := fmt.Sprintf("%s.%s%s", userIdHash, hash, ext)
			list = append(list, &photoInfo{Name: name, File: file})
			i++
		}
	}

	err := saveFiles(ctx, env, list)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(list))
	for i := 0; i < len(list); i++ {
		names[i] = list[i].Name
	}

	return names, nil
}

func saveFiles(ctx context.Context, env *env.Environment, files []*photoInfo) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, file := range files {
		path := filepath.Join(env.Settings.StaticStorage.Path, file.Name)
		eg.Go(func() error {
			return saveFile(ctx, path, file.File)
		})
	}

	return eg.Wait()
}

func saveFile(ctx context.Context, path string, file *multipart.FileHeader) (err error) {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		errClose := dst.Close()
		if err == nil {
			err = errClose
		}
	}()

	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(dst, f)
	if err != nil {
		return err
	}

	return nil
}

func removeFilesByUrl(path string, urls []string) error {
	for _, url := range urls {
		photoPath := filepath.Join(path, filepath.Base(url))
		_ = os.Remove(photoPath)
	}

	return nil
}
