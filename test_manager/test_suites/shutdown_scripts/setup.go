package shutdown_scripts

import (
	"fmt"

	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_manager"
)

var (
	Name           = "shutdown-scripts"
	shutdownscript = `#!/bin/bash
count=1
while True; do
  echo $count | tee -a /root/the_log
  ((count++)
  sleep 1
done`
)

func TestSetup(t *test_manager.TestSuite) error {
	fmt.Println("shutdown-scripts.TestSetup")
	vm1, err := t.CreateTestVM("vm")
	if err != nil {
		return err
	}
	vm1.SetShutdownScript(shutdownscript)
	return nil
	// Do I then register this VM, or can CreateVM really be smart enough to figure out what I mean?
	//
	// or another example
	// vm1 := test_helpers.CreateVM()
	// vm2 := test_helpers.CreateVM()
	// vm1.RunTests("TestXxx,TestZzz")
	// vm2.RunTests("TestYyy")
	//
	// or vm1.RunTests()
	// or we can omit that, and it just sets a filter (ultimately a metadata key)
	//
	// skip might look like:
	// test_helpers.AddSkipImages("ubuntu-*")  // these tests can't run on ubuntu
}
