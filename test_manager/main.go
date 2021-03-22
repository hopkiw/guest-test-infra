package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_manager"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/image_validation"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/oslogin"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/shutdown_scripts"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/ssh"
)

var (
	project       = flag.String("project", "", "project to be used for tests")
	zone          = flag.String("zone", "", "zone to be used for tests")
	print         = flag.Bool("print", false, "print out the parsed test workflows and exit")
	validate      = flag.Bool("validate", false, "validate all the test workflows and exit")
	filter        = flag.String("filter", "", "test suite name filter")
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
// TODO: we need to figure out logging, skips, failures in testsetup, etc.

// main alone has the packages imported and can reference their TestSetup
// functions.. even if it's only to register them.
func main() {
	ctx := context.Background()

	// Setup logging.

	testLogger := log.New(os.Stdout, "[TestManager] ", 0)
	testLogger.Println("Starting...")

	// TODO: this was copied from osconfig tests. Do we use anything that needs this?
	//       I think this is for if any shared code does logging, we
	//       control its format. I don't think we invoke any code that uses
	//       guest-logging-go here, though.
	/*
		opts := logger.LogOpts{LoggerName: "TestManager-cl", Debug: true,
			Writers: []io.Writer{&logWriter{log: testLogger}}, DisableCloudLogging: true, DisableLocalLogging: true}
		logger.Init(ctx, opts)
	*/

	// Handle args.

	flag.Parse()
	if *project == "" || *zone == "" || *images == "" {
		testLogger.Fatal("Must provide project, zone and images arguments")
		return
	}

	// Setup tests.

	testPackages := []struct {
		name      string
		setupFunc func(*test_manager.TestSuite) error
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

	var testSuites []*test_manager.TestSuite
	for _, testPackage := range testPackages {
		for _, image := range strings.Split(*images, ",") {
			ts := &test_manager.TestSuite{Name: testPackage.name, Image: image}
			testSuites = append(testSuites, ts)
			if err := testPackage.setupFunc(ts); err != nil {
				testLogger.Printf("%s.TestSetup for %s failed: %v", testPackage.name, image, err)
				ts.Disable()
			}
		}
	}

	testLogger.Println("test_manager: Done with setup!")
	test_manager.RunTests(ctx, testSuites, testLogger, *outPath, *project, *zone, *parallelCount)
}
