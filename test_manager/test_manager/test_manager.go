package test_manager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GoogleCloudPlatform/compute-image-tools/daisy"
)

type TestSuite struct {
	wf    *daisy.Workflow
	Name  string
	Image string
}

func (t *TestSuite) String() string {
	b, err := json.MarshalIndent(t.wf, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling workflow for printing:", err)
	}
	return string(b)
}

func (t *TestSuite) Disable() {
	t.wf = nil
}

// RunTests runs a set of TestSuites as one TestRun. Workflows are finalized
// and then run.
func RunTests(tests []*TestSuite) error {
	for _, ts := range tests {
		if ts.wf == nil {
			fmt.Printf("warn: no workflow for %s\n", ts.Name)
			continue
		}
		ts.wf.Sources["startup"] = "/test_manager/test_wrapper/test_wrapper"
		ts.wf.Sources["testbinary"] = fmt.Sprintf("/test_manager/test_suites/%s/%s.test", ts.Name, ts.Name)
		for _, vm := range ts.wf.Steps["create-vms"].CreateInstances.Instances {
			if vm.Metadata == nil {
				vm.Metadata = make(map[string]string)
			}
			vm.Metadata["_test_testbinarypath"] = "${SOURCESPATH}/testbinary"
		}

		// TODO : parallelize.
		fmt.Printf("Running test %s on image %s\n", ts.Name, ts.Image)
		ts.wf.Print(context.Background())
		/*
			if err := ts.wf.Run(context.Background()); err != nil {
				return err
			}
		*/
	}
	return nil
}
