package oslogin

import (
	"errors"

	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/testmanager"
)

var Name = "oslogin"

func TestSetup(t *testmanager.TestWorkflow) error {
	if t.Image == "centos-7" {
		return errors.New("dummy error")
	}
	return testmanager.SingleVMTest(t)
}
