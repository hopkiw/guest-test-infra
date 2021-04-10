package image_validation

import "github.com/GoogleCloudPlatform/guest-test-infra/cloud_image_tests/testmanager"

var Name = "image-validation"

func TestSetup(t *testmanager.TestWorkflow) error {
	return testmanager.SingleVMTest(t)
}
