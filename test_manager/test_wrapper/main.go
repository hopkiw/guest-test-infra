package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/utils"
	junitFormatter "github.com/jstemmer/go-junit-report/formatter"
	junitParser "github.com/jstemmer/go-junit-report/parser"
)

const (
	testResult          = "junit.xml"
	testBinaryLocalPath = "image_test.test"
)

func main() {
	ctx := context.Background()
	log.Printf("started test_wrapper")

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create cloud storage client: %v\n", err)
	}

	daisyOutsPath, err := utils.GetMetadataAttribute("daisy-outs-path")
	if err != nil {
		log.Fatalf("couldnt determine daisy-outs-path: %v", err)
		return
	}
	log.Printf("daisy-outs-path: %s", daisyOutsPath)

	testRun, err := utils.GetMetadataAttribute("_test_run")
	if err != nil {
		log.Printf("couldnt determine _test_run: %v", err)
		testRun = ""
	}
	log.Printf("_test_run: %s", testRun)

	testBinaryURL, err := utils.GetMetadataAttribute("_test_binarypath")
	if err != nil {
		log.Fatalf("couldnt determine _test_binarypath: %v", err)
		return
	}
	log.Printf("_test_binarypath: %s", testBinaryURL)

	var testArguments = []string{"-test.v"}
	if testRun != "" {
		testArguments = append(testArguments, "-test.run", testRun)
	}
	log.Printf("test arguments: %v", testArguments)

	workDir, err := mkWorkDir()
	if err != nil {
		log.Fatalf("couldnt make working dir: %v", err)
		return
	}
	log.Printf("working dir: %v", workDir)

	if err := downloadGCSObject(ctx, client, testBinaryURL, workDir+testBinaryLocalPath); err != nil {
		log.Fatalf("Failed to download object %s: %v\n", testBinaryURL, err)
		return
	}

	if err := os.Chdir(workDir); err != nil {
		log.Fatalf("couldnt cd to working dir: %v", err)
		return
	}

	testOutput, err := exec.Command(workDir+testBinaryLocalPath, testArguments...).Output()
	if err != nil {
		log.Printf("command returned error or failed to run: %v", err)
	}
	log.Printf("got test output:\n%s\n", testOutput)

	testData, err := convertToJunit(string(testOutput))
	if err != nil {
		log.Fatalf("Failed to convert to junit format: %v\n", err)
		return
	}
	log.Printf("junit formatted output:\n%s\n", testData)

	if err := uploadResult(ctx, client, daisyOutsPath+"/"+testResult, testData); err != nil {
		log.Fatalf("Failed to upload test result: %v\n", err)
	}
	log.Printf("uploaded result to %s", daisyOutsPath+"/"+testResult)
	log.Printf("MAGIC-STRING: done!")
}

func mkWorkDir() (string, error) {
	workDir := "/test-" + randString(5)
	if err := os.Mkdir(workDir, 0755); err != nil {
		return "", err
	}
	return workDir, nil
}

func convertToJunit(input string) (*bytes.Buffer, error) {
	var b bytes.Buffer
	report, err := junitParser.Parse(strings.NewReader(input), "")
	if err != nil {
		return nil, err
	}
	if err = junitFormatter.JUnitReportXML(report, false, "", &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func downloadGCSObject(ctx context.Context, client *storage.Client, gcsPath, file string) error {
	u, err := url.Parse(gcsPath)
	if err != nil {
		log.Fatalf("Failed to parse GCS url: %v\n", err)
	}
	rc, err := client.Bucket(u.Host).Object(u.Path[1:]).NewReader(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(file, data, 0755); err != nil {
		return err
	}
	return nil
}

func uploadResult(ctx context.Context, client *storage.Client, path string, data io.Reader) error {
	u, err := url.Parse(path)
	if err != nil || u.Path == "" {
		log.Fatalf("Failed to parse URL: %v\n", err)
	}
	des := client.Bucket(u.Host).Object(u.Path[1:]).NewWriter(ctx)
	if _, err := io.Copy(des, data); err != nil {
		return fmt.Errorf("Failed to write file: %v", err)
	}
	des.Close()
	return nil
}

func randString(n int) string {
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	letters := "bdghjlmnpqrstvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[gen.Int63()%int64(len(letters))]
	}
	return string(b)
}
