package image_validation

import "github.com/GoogleCloudPlatform/guest-test-infra/imagetest"

var Name = "image-validation"

func TestSetup(t *imagetest.TestWorkflow) error {
	return imagetest.SingleVMTest(t)
}
