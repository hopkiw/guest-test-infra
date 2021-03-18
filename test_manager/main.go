package main

import (
	"fmt"

	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_manager"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/image_validation"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/oslogin"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/shutdown_scripts"
	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_suites/ssh"
)

func main() {
	// Normally this would be provided as an argument.
	image := "projects/debian-cloud/global/images/family/debian-10"

	fmt.Println("test_manager: creating shutdown scripts test suite")
	shutdown_scripts_testsuite := &test_manager.TestSuite{
		Name:  shutdown_scripts.Name,
		Image: image,
	}
	if err := shutdown_scripts.TestSetup(shutdown_scripts_testsuite); err != nil {
		fmt.Printf("Got an error setting up %s: %v\n", shutdown_scripts_testsuite.Name, err)
		shutdown_scripts_testsuite.Disable()
	}

	fmt.Println("test_manager: creating image validation test suite")
	image_validation_testsuite := &test_manager.TestSuite{
		Name:  image_validation.Name,
		Image: image,
	}
	if err := image_validation.TestSetup(image_validation_testsuite); err != nil {
		fmt.Printf("Got an error setting up %s: %v\n", image_validation_testsuite.Name, err)
		image_validation_testsuite.Disable()
	}

	fmt.Println("test_manager: creating oslogin test suite")
	oslogin_testsuite := &test_manager.TestSuite{
		Name:  oslogin.Name,
		Image: image,
	}
	if err := oslogin.TestSetup(oslogin_testsuite); err != nil {
		fmt.Printf("Got an error setting up %s: %v\n", oslogin_testsuite.Name, err)
		oslogin_testsuite.Disable()
	}

	fmt.Println("test_manager: creating ssh test suite")
	ssh_testsuite := &test_manager.TestSuite{
		Name:  ssh.Name,
		Image: image,
	}
	if err := ssh.TestSetup(ssh_testsuite); err != nil {
		fmt.Printf("Got an error setting up %s: %v\n", ssh_testsuite.Name, err)
		ssh_testsuite.Disable()
	}

	fmt.Println("test_manager: Done with setup!")

	test_manager.RunTests([]*test_manager.TestSuite{shutdown_scripts_testsuite, image_validation_testsuite, oslogin_testsuite, ssh_testsuite})

}

/*
func main() {
	ts := &test_manager.TestSuite{Name: "fake", Image: "projects/debian-cloud/global/images/family/debian-10"}
	// here's testsetup
	vm, err := ts.CreateTestVM("testvm")
	if err != nil {
		panic(err)
	}
	vm.AddMetadata("mykey", "myvalue")
	if err := test_manager.RunTests([]*test_manager.TestSuite{ts}); err != nil {
		panic(err)
	}
}
*/
