package test_manager

import (
	"fmt"

	"github.com/GoogleCloudPlatform/compute-image-tools/daisy"
	"google.golang.org/api/compute/v1"
)

// SingleVMTest configures the simple test case of one VM running the test
// suite.
func SingleVMTest(t *TestSuite) error {
	_, err := t.CreateTestVM(t.Name)
	return err
}

// TestVM is a test VM.
type TestVM struct {
	name      string
	testSuite *TestSuite
}

// CreateTestVM adds steps to a workflow to create a test VM. The workflow is
// created if it doesn't exist. The first VM created has a WaitInstances step
// configured.
func (t *TestSuite) CreateTestVM(name string) (*TestVM, error) {
	name = "testvm"
	testVM := &TestVM{name: name, testSuite: t}

	if t.wf == nil {
		t.wf = daisy.New()
		t.wf.Name = t.Name
		t.wf.Project = "liamh-testing"
		t.wf.Zone = "us-west1-b"
		if err := t.createFirstTestVM(name); err != nil {
			return nil, err
		}
		return testVM, nil
	}

	// Test VMs and disks are created in parallel, so we're just appending.
	bootdisk := &daisy.Disk{}
	bootdisk.Name = name
	bootdisk.SourceImage = t.Image
	createDisksStep := t.wf.Steps["create-disks"]
	*createDisksStep.CreateDisks = append(*createDisksStep.CreateDisks, bootdisk)

	instance := &daisy.Instance{}
	instance.StartupScript = "startup"
	instance.Name = name
	instance.Scopes = append(instance.Scopes, "https://www.googleapis.com/auth/devstorage.read_write")

	attachedDisk := &compute.AttachedDisk{Source: name}
	instance.Disks = append(instance.Disks, attachedDisk)

	createVMStep := t.wf.Steps["create-vms"]
	createVMStep.CreateInstances.Instances = append(createVMStep.CreateInstances.Instances, instance)

	return testVM, nil
}

func (t *TestSuite) createFirstTestVM(name string) error {
	bootdisk := &daisy.Disk{}
	bootdisk.Name = name
	bootdisk.SourceImage = t.Image
	createDisks := &daisy.CreateDisks{bootdisk}
	createDisksStep, err := t.wf.NewStep("create-disks")
	if err != nil {
		return err
	}
	createDisksStep.CreateDisks = createDisks

	instance := &daisy.Instance{}
	instance.StartupScript = "startup"
	instance.Name = name
	instance.Scopes = append(instance.Scopes, "https://www.googleapis.com/auth/devstorage.read_write")

	attachedDisk := &compute.AttachedDisk{Source: name}
	instance.Disks = append(instance.Disks, attachedDisk)

	createInstances := &daisy.CreateInstances{}
	createInstances.Instances = append(createInstances.Instances, instance)

	createVMStep, err := t.wf.NewStep("create-vms")
	if err != nil {
		return err
	}
	_ = t.wf.Steps["create-vms"]
	createVMStep.CreateInstances = createInstances
	if err := t.wf.AddDependency(createVMStep, createDisksStep); err != nil {
		return err
	}

	instanceSignal := &daisy.InstanceSignal{}
	instanceSignal.Name = name
	instanceSignal.Stopped = true
	waitForInstances := &daisy.WaitForInstancesSignal{instanceSignal}

	waitStep, err := t.wf.NewStep("wait-" + name)
	if err != nil {
		return err
	}
	waitStep.WaitForInstancesSignal = waitForInstances

	t.wf.AddDependency(waitStep, createVMStep)

	return nil
}

// AddMetadata adds the specified key:value pair to metadata during VM creation.
func (t *TestVM) AddMetadata(key, value string) {
	createVMStep := t.testSuite.wf.Steps["create-vms"]
	for _, vm := range createVMStep.CreateInstances.Instances {
		if vm.Name == t.name {
			if vm.Metadata == nil {
				vm.Metadata = make(map[string]string)
			}
			vm.Metadata[key] = value
			return
		}
	}
}

// RunTests runs only the named tests on the testVM. `runtest` must match the
// parameter to `go test -run`
func (t *TestVM) RunTests(runtest string) {
	t.AddMetadata("_test_run", runtest)
}

// SetShutdownScript sets the `shutdown-script` metadata key for a VM.
func (t *TestVM) SetShutdownScript(script string) {
	t.AddMetadata("shutdown-script", script)
}

// SetStartupScript sets the `startup-script` metadata key for a VM.
func (t *TestVM) SetStartupScript(script string) {
	t.AddMetadata("startup-script", script)
}

// AddWait adds a daisy.WaitForInstancesSignal Step depending on this VM. Note:
// the first created VM automatically has a wait step created. It is an error
// to call this function on the first VM in a workflow.
func (t *TestVM) AddWait(success, failure, status string, stopped bool) error {
	if _, ok := t.testSuite.wf.Steps["wait-"+t.name]; ok {
		return fmt.Errorf("wait step already exists for TestVM %q", t.name)
	}
	instanceSignal := &daisy.InstanceSignal{}
	instanceSignal.Name = t.name
	instanceSignal.Stopped = stopped
	if success != "" {
		instanceSignal.SerialOutput.SuccessMatch = success
	}
	if status != "" {
		instanceSignal.SerialOutput.StatusMatch = status
	}
	if failure != "" {
		// FailureMatch is a []string, compared to success and status.
		instanceSignal.SerialOutput.FailureMatch = append(instanceSignal.SerialOutput.FailureMatch, failure)
	}
	waitForInstances := &daisy.WaitForInstancesSignal{instanceSignal}
	s, err := t.testSuite.wf.NewStep("wait-" + t.name)
	if err != nil {
		return err
	}
	s.WaitForInstancesSignal = waitForInstances
	t.testSuite.wf.AddDependency(s, t.testSuite.wf.Steps["create-vms"])
	return nil
}
