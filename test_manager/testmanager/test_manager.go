package test_manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/compute-image-tools/daisy"
)

var (
	client *storage.Client
)

const (
	testBinariesPath = "/tmp/test_manager"
	testWrapperPath  = testBinariesPath + "/test_wrapper/test_wrapper"
	testSuitePath    = testBinariesPath + "/test_suites"
)

type TestSuite struct {
	wf      *daisy.Workflow
	Name    string
	Image   string
	skipped bool
}

// For testing
func (t *TestSuite) String() string {
	b, err := json.MarshalIndent(t.wf, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling workflow for printing:", err)
	}
	return string(b)
}

// For testing
func (t *TestSuite) Disable() {
	t.wf = nil
}

func finalizeWorkflows(tests []*TestSuite, testLogger *mylogger, zone, project string) {
	for _, ts := range tests {
		if ts.wf == nil {
			continue
		}

		ts.wf.DisableGCSLogging()
		ts.wf.DisableCloudLogging()
		ts.wf.DisableStdoutLogging()
		ts.wf.Logger = testLogger

		ts.wf.Zone = zone
		ts.wf.Project = project

		ts.wf.Sources["startup"] = testWrapperPath
		ts.wf.Sources["testbinary"] = fmt.Sprintf("%s/%s/%s.test", testSuitePath, ts.Name, ts.Name)

		// Add metadata to each VM.
		for _, vm := range ts.wf.Steps["create-vms"].CreateInstances.Instances {
			if vm.Metadata == nil {
				vm.Metadata = make(map[string]string)
			}
			vm.Metadata["_test_binarypath"] = "${SOURCESPATH}/testbinary"
		}
	}
}

func downloadGCSObject(ctx context.Context, client *storage.Client, gcsPath string) (string, error) {
	fmt.Printf("  downloadGCSObject: trying to get %s\n", gcsPath)
	u, err := url.Parse(gcsPath)
	if err != nil {
		log.Fatalf("Failed to parse GCS url: %v\n", err)
	}
	rc, err := client.Bucket(u.Host).Object(u.Path[1:]).NewReader(ctx)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type TestResult struct {
	testSuite                       *TestSuite
	Skipped, FailedSetup            bool
	WorkflowFailed, WorkflowSuccess bool
	Result                          string
}

func getTestResults(ctx context.Context, test *TestSuite) (string, error) {
	// TODO: *tell* the test where to put results. never infer it with
	//       hardcoded logic like 'the gcs path + outs + junit.xml'
	gcsPath := "gs://liamh-export/test_new_test_manager/0001/" + test.Name + "/junit.xml"
	junitString, err := downloadGCSObject(ctx, client, gcsPath)
	if err != nil {
		return "", err
	}

	return junitString, nil
}

func runTestSuite(ctx context.Context, test *TestSuite) TestResult {
	var res TestResult
	res.testSuite = test
	if test.skipped {
		res.Skipped = true
		return res
	}
	if test.wf == nil {
		res.FailedSetup = true
		return res
	}
	fmt.Printf("runTestSuite: running %s on %s (ID %s)\n%s\n", test.Name, test.Image, test.wf.ID(), test.String())
	if err := test.wf.Run(ctx); err != nil {
		fmt.Printf("returning error: %v\n", err)
		res.WorkflowFailed = true
		res.Result = err.Error()
		return res
	}
	results, err := getTestResults(ctx, test)
	if err != nil {
		res.WorkflowFailed = true
		res.Result = err.Error()
		return res
	}
	res.WorkflowSuccess = true
	res.Result = results

	return res
}

// === shamelessly stolen from daisy_test_runner to test ===

type mylogger struct {
	buf bytes.Buffer
	mx  sync.Mutex
}

func (l *mylogger) WriteLogEntry(e *daisy.LogEntry) {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.buf.WriteString(e.String())
}

func (l *mylogger) WriteSerialPortLogs(w *daisy.Workflow, instance string, buf bytes.Buffer) {
	return
}

func (l *mylogger) ReadSerialPortLogs() []string {
	return nil
}

func (l *mylogger) Flush() { return }

// === END stolen ===

func RunTests(ctx context.Context, testSuites []*TestSuite, testLogger *log.Logger, outPath, project, zone string, parallelCount int) {
	var err error
	client, err = storage.NewClient(ctx)
	if err != nil {
		testLogger.Fatalf("Failed to set up storage client: %v", err)
	}
	// seems testLogger is unused in this model.
	finalizeWorkflows(testSuites, &mylogger{}, zone, project)

	testResults := make(chan TestResult, len(testSuites))
	tests := make(chan *TestSuite, len(testSuites))

	var wg sync.WaitGroup
	for i := 0; i < parallelCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for test := range tests {
				testResults <- runTestSuite(ctx, test)
			}
			fmt.Printf("goroutine %d exiting\n", id)
		}(i)
	}
	for _, item := range testSuites {
		tests <- item
	}
	fmt.Println("done adding tests to work chan, closing it")
	close(tests)
	fmt.Println("waiting on waitgroup")
	wg.Wait()
	fmt.Println("waitgroup done, getting test results from channel")
	for i := 0; i < len(testSuites); i++ {
		res := <-testResults
		fmt.Printf("got a test result for %s: %+v\n", res.testSuite.Name, res)
	}
	fmt.Println("all done!")
}
