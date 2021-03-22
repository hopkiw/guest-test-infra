package testmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/compute-image-tools/daisy"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/utils"
)

var (
	client *storage.Client
)

const (
	testBinariesPath = "/tmp/test_manager"
	testWrapperPath  = testBinariesPath + "/test_wrapper/test_wrapper"
	testSuitePath    = testBinariesPath + "/test_suites"
)

// Originally named TestSuite because a jUnit <testsuite> will cover one workflow.

// TestWorkflow defines a test workflow which creates at least one test VM.
type TestWorkflow struct {
	wf          *daisy.Workflow
	Name        string
	Image       string
	skipped     bool
	destination string
}

// TODO: remove this testing method
func (t *TestWorkflow) String() string {
	b, err := json.MarshalIndent(t.wf, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling workflow for printing:", err)
		return ""
	}
	return string(b)
}

// TODO: remove this testing method
func (t *TestWorkflow) Disable() {
	t.wf = nil
}

func finalizeWorkflows(tests []*TestWorkflow, zone, project string) {
	for _, ts := range tests {
		if ts.wf == nil {
			continue
		}
		ts.destination = fmt.Sprintf("gs://liamh-export/new_test_manager_testing/0001/%s-%s/junit.xml", ts.wf.Name, ts.wf.ID())

		ts.wf.DisableGCSLogging()
		ts.wf.DisableCloudLogging()
		ts.wf.DisableStdoutLogging()

		ts.wf.Zone = zone
		ts.wf.Project = project

		ts.wf.Sources["startup"] = testWrapperPath
		// TODO: the name i know is maybe different from path name
		ts.wf.Sources["testbinary"] = fmt.Sprintf("%s/%s/%s.test", testSuitePath, ts.Name, ts.Name)

		copyStep := ts.wf.Steps["copy-objects"]
		// Two issues with manipulating this step. First, it is a
		// typedef that removes the slice notation, so we have to cast
		// it back in order to index it.
		copyObject := []daisy.CopyGCSObject(*copyStep.CopyGCSObjects)[0]
		copyObject.Destination = ts.destination
		// Second, it is not a pointer, so we can't modify it in place.
		// Instead, we overwrite the struct with a new step with our
		// modified copy of the config.
		copyStep.CopyGCSObjects = &daisy.CopyGCSObjects{copyObject}

		// Add metadata to each VM.
		for _, vm := range ts.wf.Steps["create-vms"].CreateInstances.Instances {
			if vm.Metadata == nil {
				vm.Metadata = make(map[string]string)
			}
			vm.Metadata["_test_binarypath"] = "${SOURCESPATH}/testbinary"
		}
	}
}

type TestResult struct {
	testWorkflow                    *TestWorkflow
	Skipped, FailedSetup            bool
	WorkflowFailed, WorkflowSuccess bool
	Result                          string
}

func getTestResults(ctx context.Context, ts *TestWorkflow) (string, error) {
	junit, err := utils.DownloadGCSObject(ctx, client, ts.destination)
	if err != nil {
		return "", err
	}

	return string(junit), nil
}

func runTestWorkflow(ctx context.Context, test *TestWorkflow) TestResult {
	var res TestResult
	res.testWorkflow = test
	if test.skipped {
		res.Skipped = true
		return res
	}
	if test.wf == nil {
		res.FailedSetup = true
		return res
	}
	// TODO: remove this debug line
	// only now do i notice i got the ID before populate..
	fmt.Printf("runTestWorkflow: running %s on %s (ID %s)\n", test.Name, test.Image, test.wf.ID())
	test.wf.Print(ctx)
	/*
		if err := test.wf.Run(ctx); err != nil {
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
	*/
	res.WorkflowSuccess = true
	res.Result = "results"

	return res
}

func PrintTests(ctx context.Context, testWorkflows []*TestWorkflow, project, zone string) {
	finalizeWorkflows(testWorkflows, zone, project)
	for _, test := range testWorkflows {
		test.wf.Print(ctx)
	}
}

func ValidateTests(ctx context.Context, testWorkflows []*TestWorkflow, project, zone string) error {
	finalizeWorkflows(testWorkflows, zone, project)
	for _, test := range testWorkflows {
		if err := test.wf.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

func RunTests(ctx context.Context, testWorkflows []*TestWorkflow, outPath, project, zone string, parallelCount int) {
	var err error
	client, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to set up storage client: %v", err)
	}
	finalizeWorkflows(testWorkflows, zone, project)

	testResults := make(chan TestResult, len(testWorkflows))
	testchan := make(chan *TestWorkflow, len(testWorkflows))

	var wg sync.WaitGroup
	for i := 0; i < parallelCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for test := range testchan {
				testResults <- runTestWorkflow(ctx, test)
			}
		}(i)
	}
	for _, ts := range testWorkflows {
		testchan <- ts
	}
	close(testchan)
	wg.Wait()
	for i := 0; i < len(testWorkflows); i++ {
		//	res := <-testResults
		parseResult(<-testResults)
	}
}

func parseResult(res TestResult) {
	switch {
	case res.FailedSetup:
		fmt.Printf("test %s on %s failed during setup and was disabled\n", res.testWorkflow.Name, res.testWorkflow.Image)
	case res.Skipped:
		fmt.Printf("test %s on %s was skipped\n", res.testWorkflow.Name, res.testWorkflow.Image)
	case res.WorkflowFailed:
		fmt.Printf("test %s on %s worklow failed: %s\n", res.testWorkflow.Name, res.testWorkflow.Image, res.Result)
	case res.WorkflowSuccess:
		fmt.Printf("test %s on %s passed\n", res.testWorkflow.Name, res.testWorkflow.Image)
	default:
		fmt.Printf("test %s on %s has unknown status\n", res.testWorkflow.Name, res.testWorkflow.Image)
	}
}
