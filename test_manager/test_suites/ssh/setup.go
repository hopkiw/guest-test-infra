package ssh

import (
	"fmt"

	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_manager"
)

var Name = "ssh"

func TestSetup(t *test_manager.TestSuite) error {
	fmt.Println("ssh.TestSetup")
	vm1, _ := t.CreateTestVM("vm1")
	vm1.RunTests("TestVm1")
	vm2, _ := t.CreateTestVM("vm2")
	vm2.RunTests("TestVm2")
	return nil
}
