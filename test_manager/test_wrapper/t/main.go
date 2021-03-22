package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"cloud.google.com/go/storage"
)

var (
	file    = "gs://liamh-testing-daisy-bkt/daisy-ssh-20210321-23:25:22-6dhx4/sources/testbinary"
	outpath = "/tmp/outpath"
)

func main() {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create cloud storage client: %v\n", err)
	}
	if err := downloadGCSObject(ctx, client, file, outpath); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("success! wrote to %s\n", outpath)
}

func downloadGCSObject(ctx context.Context, client *storage.Client, gcsPath, file string) error {
	u, err := url.Parse(gcsPath)
	if err != nil {
		log.Fatalf("Failed to parse GCS url: %v\n", err)
	}
	bucket, object := u.Host, u.Path
	fmt.Printf("  downloadGCSObject: bucket %q object %q\n", bucket, object)
	fmt.Printf("  downloadGCSObject: confirm %q == %q\n", gcsPath, "gs://"+u.Host+u.Path)
	rc, err := client.Bucket(bucket).Object(object[1:]).NewReader(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()
	fmt.Printf("  downloadGCSObject: got object\n")

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(file, data, 0755); err != nil {
		return err
	}
	return nil
}
