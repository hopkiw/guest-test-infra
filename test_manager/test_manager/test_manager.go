package test_manager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GoogleCloudPlatform/compute-image-tools/daisy"
	"google.golang.org/api/compute/v1"
)

type TestSuite struct {
	wf    *daisy.Workflow
	Name  string
	Image string
}

type TestVM struct {
	name      string
	testSuite *TestSuite
}

func (t *TestSuite) String() string {
	b, err := json.MarshalIndent(t.wf, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling workflow for printing:", err)
	}
	return string(b)
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

func (t *TestSuite) Disable() {
	t.wf = nil
}

// AddMetadata adds the specified key:value pair to metadata during VM creation.
func (t *TestVM) AddMetadata(key, value string) {
	step := t.testSuite.wf.Steps["create-vms"]
	//for _, vm := range t.testSuite.wf.Steps["create-vms"].CreateInstances.Instances {
	for _, vm := range step.CreateInstances.Instances {
		if vm.Name == t.name {
			if vm.Metadata == nil {
				vm.Metadata = make(map[string]string)
			}
			vm.Metadata[key] = value
			//			step.CreateInstances.Instances[idx] = vm
			return
		}
	}
}

// RunTests runs only the named tests on the testVM. `runtest` must match the
// parameter to `go test -run`
func (t *TestVM) RunTests(runtest string) {
	t.AddMetadata("test.run", runtest)
}

func (t *TestVM) SetShutdownScript(script string) {
	t.AddMetadata("shutdown-script", script)
}

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
	if failure != "" {
		instanceSignal.SerialOutput.FailureMatch = append(instanceSignal.SerialOutput.FailureMatch, failure)
	}
	if status != "" {
		instanceSignal.SerialOutput.StatusMatch = status
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

func RunTests(tests []*TestSuite) error {
	for _, ts := range tests {
		if ts.wf == nil {
			fmt.Printf("warn: no workflow for %s\n", ts.Name)
			continue
		}
		ts.wf.Sources["startup"] = "/tmp/test_manager/test_wrapper/test_wrapper"
		ts.wf.Sources["testbinary"] = fmt.Sprintf("/tmp/test_manager/test_suites/%s/%s.test", ts.Name, ts.Name)
		for _, vm := range ts.wf.Steps["create-vms"].CreateInstances.Instances {
			if vm.Metadata == nil {
				vm.Metadata = make(map[string]string)
			}
			vm.Metadata["test.testbinarypath"] = "${SOURCESPATH}/testbinary"
		}

		fmt.Printf("Running test %s on image %s\n", ts.Name, ts.Image)
		//		fmt.Printf("Workflow:\n%s\n", ts.String())
		ts.wf.Print(context.Background())
		/*
			if err := ts.wf.Run(context.Background()); err != nil {
				return err
			}
		*/
	}
	return nil
}

// utils

func SingleVMTest(t *TestSuite) error {
	// This is a simple TestSetup function for convenience sake.
	_, err := t.CreateTestVM(t.Name)
	return err
}
