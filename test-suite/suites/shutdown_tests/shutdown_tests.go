package shutdown_tests

import "fmt"

// This is where you write setup code, override defaults, etc.
// You could maybe have a return value, or just optional package hooks for registering things?
// Are you required to provide it, or do we use reflection to find out and if not, invoke a default?
func setup() {
	fmt.Printf("shutdown_tests: setup()")
	//vm1 := test_helpers.CreateVM()
	//vm1.SetShutdownScript(shutdownscript)
	// Do I then register this VM, or can CreateVM really be smart enough to figure out what I mean?
	//
	// or another example
	// vm1 := test_helpers.CreateVM()
	// vm2 := test_helpers.CreateVM()
	// vm1.RunTests("TestXxx,TestZzz")
	// vm2.RunTests("TestYyy")
	//
	// or vm1.RunTests()
	// or we can omit that, and it just sets a filter (ultimately a metadata key)
}
