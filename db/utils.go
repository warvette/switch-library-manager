package db

import (
	bytes2 "bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func LoadAndUpdateFile(url string, filePath string, etag string) (*os.File, string, error) {

	//load file
	file, err := os.Open(filePath)

	//try to check if there is a new version
	//if so, save the file
	bytes, newEtag, err := downloadBytesFromUrl(url, etag)
	if err == nil {
		//validate json structure
		var test map[string]interface{}
		err = decodeToJsonObject(bytes2.NewReader(bytes), &test)
		if err == nil {
			file, err = saveFile(bytes, filePath)
			etag = newEtag
		} else {
			fmt.Printf("mailformed json file - [%v], error: [%v]\n", url, err)
		}
	} else {
		fmt.Printf("file [%v] was not downloaded, reason - [%v]\n", url, err)
	}

	return file, etag, nil
}

func decodeToJsonObject(reader io.Reader, target interface{}) error {
	err := json.NewDecoder(reader).Decode(target)
	if err != nil {
		return err
	}
	return nil
}

func downloadBytesFromUrl(url string, etag string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("If-None-Match", etag)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode >= 400 {
		return nil, "", errors.New("got a non 200 response - " + resp.Status)
	}
	defer resp.Body.Close()
	//getting the new etag
	etag = resp.Header.Get("Etag")

	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}
		return body, etag, nil
	}

	return nil, "", errors.New("no new updates")
}

func saveFile(bytes []byte, fileName string) (*os.File, error) {

	err := ioutil.WriteFile("./"+fileName, bytes, 0644)
	if err != nil {
		return nil, err
	}

	file, err := os.Open("./" + fileName)
	if err != nil {
		return nil, err
	}
	return file, nil
}