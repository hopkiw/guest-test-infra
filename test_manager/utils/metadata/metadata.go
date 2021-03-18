package metadata

import (
	"io/ioutil"
	"net/http"
)

var baseurl = "http://metadata.google.internal/computeMetadata/v1/instance"

func GetMetadata(key string) (string, int, error) {
	req, err := http.NewRequest("GET", baseurl+key, nil)
	if err != nil {
		return "", -1, err
	}
	req.Header.Add("Metadata-Flavor", "Google")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", -1, err
	}
	val, err := ioutil.ReadAll(resp.Body)
	return string(val), resp.StatusCode, err
}
