package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const metadataUrlPrefix = "http://metadata.google.internal/computeMetadata/v1/instance/attributes/"

// GetRealVMName returns the real name of a VM running in the same test.
func GetRealVMName(name string) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(hostname, "-", 3)
	if len(parts) != 3 {
		return "", errors.New("hostname doesn't match scheme")
	}
	return strings.Join([]string{parts[0], name, parts[2]}, "-"), nil
}

func GetMetadataAttribute(attribute string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", metadataUrlPrefix, attribute), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Metadata-Flavor", "Google")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("http response code is %v", resp.StatusCode)
	}
	val, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(val), nil
}
