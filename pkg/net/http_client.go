package net

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

func httpGet(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("got error from net.Get. url: %q. Error: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP GET response body reading error. url: %q. Error: %w", url, err)
	}

	if resp.StatusCode != 200 {
		return errors.New(string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("HTTP GET response body parsing error. url: %q. Error: %w", url, err)
	}
	return nil
}

func httpPost(url string) error {
	resp, err := http.Post(url, "application/text", nil)
	if err != nil {
		return fmt.Errorf("can not send POST. Network error. url: %q. Error: %w", url, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP POST response body reading error. url: %q. Error: %w", url, err)
	}

	if resp.StatusCode != 200 {
		return errors.New(string(body))
	}

	return nil
}

func HttpPostJson(
	logger logr.Logger,
	url string,
	data interface{},
) (s string, resultErr error) {

	defer func() {
		if r := recover(); r != nil {
			logger.Error(nil, "Can not send POST request", "Panic", r)
			resultErr = errors.New("panic in httpPostData")
		}
	}()

	reqBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("can not send POST. Can not convert data to bytes. url: %q. Error: %w", url, err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Info("Got error from client.Do")
		return "", fmt.Errorf("can not send POST. Got error from client.Do(). url: %q. Error: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HTTP POST response body reading error. url: %q. Error: %w", url, err)
	}

	if resp.StatusCode != 200 {
		return "", errors.New(string(body))
	}

	return string(body), nil
}

type UploadFileInfo struct {
	Path string
	Data []byte
}

func HttpPostStatic(
	logger logr.Logger,
	url string,
	paramName string,
	data []*UploadFileInfo,
) (s string, resultErr error) {

	defer func() {
		if r := recover(); r != nil {
			logger.Error(nil, "Can not send POST request", "Panic", r)
			resultErr = errors.New("panic in httpPostData")
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, file := range data {
		part, err := writer.CreateFormFile(paramName, file.Path)
		if err != nil {
			return "", err
		}

		reader := bytes.NewReader(file.Data)
		_, err = io.Copy(part, reader)

		if err != nil {
			return "", err
		}
	}

	err := writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Info("Got error from client.Do")
		return "", fmt.Errorf("can not send POST. Got error from client.Do(). url: %q. Error: %w", url, err)
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("HTTP POST response body reading error. url: %q. Error: %w", url, err)
	}

	if resp.StatusCode != 200 {
		return "", errors.New(string(result))
	}

	return string(result), nil
}
