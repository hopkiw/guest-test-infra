package image_validation

import (
	"fmt"

	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_manager"
)

var Name = "image-validation"

func TestSetup(t *test_manager.TestSuite) error {
	fmt.Println("image-validation.TestSetup")
	return test_manager.SingleVMTest(t)
}
