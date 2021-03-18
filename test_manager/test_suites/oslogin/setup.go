package oslogin

import (
	"errors"
	"fmt"

	"github.com/GoogleCloudPlatform/guest-test-infra/test_manager/test_manager"
)

var Name = "oslogin"

func TestSetup(t *test_manager.TestSuite) error {
	fmt.Println("oslogin.TestSetup")
	return errors.New("dummy error")
}
