package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/image_validation"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/oslogin"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/shutdown_scripts"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/ssh"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/testmanager"
)

var (
	project       = flag.String("project", "", "project to be used for tests")
	zone          = flag.String("zone", "", "zone to be used for tests")
	printwf       = flag.Bool("print", false, "print out the parsed test workflows and exit")
	validate      = flag.Bool("validate", false, "validate all the test workflows and exit")
	filter        = flag.String("filter", "", "test name filter")
	outPath       = flag.String("out_path", "junit.xml", "junit xml path")
	images        = flag.String("images", "", "comma separated list of images to test")
	parallelCount = flag.Int("parallel_count", 5, "TestParallelCount")
)

type logWriter struct {
	log *log.Logger
}

func (l *logWriter) Write(b []byte) (int, error) {
	l.log.Print(string(b))
	return len(b), nil
}

// TODO: we need to marshall the final result of a run into a junitTestSuites
//       object with summary values
//
// TODO: we need to figure out logging, skips, failures in testsetup, etc.
//

func main() {
	flag.Parse()
	if *project == "" || *zone == "" || *images == "" {
		log.Fatal("Must provide project, zone and images arguments")
		return
	}

	// Setup tests.
	testPackages := []struct {
		name      string
		setupFunc func(*testmanager.TestWorkflow) error
	}{
		{
			image_validation.Name,
			image_validation.TestSetup,
		},
		{
			oslogin.Name,
			oslogin.TestSetup,
		},
		{
			ssh.Name,
			ssh.TestSetup,
		},
		{
			shutdown_scripts.Name,
			shutdown_scripts.TestSetup,
		},
	}

	var testWorkflows []*testmanager.TestWorkflow
	for _, testPackage := range testPackages {
		for _, image := range strings.Split(*images, ",") {
			// Would it make more sense to instantiate base
			// workflow here with a NewWorkflow function? Or is
			// there anything for the workflow we can do before
			// receiving the vm name? Will there ever be?
			// ts := testmanager.NewTestWorkflow(testPackage.Name, image)
			ts := &testmanager.TestWorkflow{Name: testPackage.name, Image: image}
			testWorkflows = append(testWorkflows, ts)
			if err := testPackage.setupFunc(ts); err != nil {
				log.Printf("%s.TestSetup for %s failed: %v", testPackage.name, image, err)
				ts.Disable()
			}
		}
	}

	log.Println("testmanager: Done with setup")

	ctx := context.Background()

	if *printwf {
		testmanager.PrintTests(ctx, testWorkflows, *project, *zone)
		return
	}

	if *validate {
		if err := testmanager.ValidateTests(ctx, testWorkflows, *project, *zone); err != nil {
			fmt.Printf("Validate failed: %v\n", err)
		}
		return
	}

	testmanager.RunTests(ctx, testWorkflows, *outPath, *project, *zone, *parallelCount)
}
