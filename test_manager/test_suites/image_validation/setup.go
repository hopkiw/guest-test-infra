package image_validation

import (
	"fmt"

	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_manager"
)

var Name = "image_validation"

func TestSetup(t *test_manager.TestSuite) error {
	fmt.Println("image_validation.TestSetup")
	return test_manager.SingleVMTest(t)
}
