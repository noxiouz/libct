package libct

import "fmt"
import "os"

func ExampleCreateCT() {
	s, err := OpenSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ct, err := s.CreateCt()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// exec
	argv := make([]string, 4)
	argv[0] = "bash"
	argv[1] = "-c"
	argv[2] = "echo"
	argv[3] = "Hello"
	env := make([]string, 0)
	err = ct.CtExecve("/bin/bash", argv, env)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// wait
	err = ct.CtWait()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Output:
	// Send:  CT_CREATE
	// Recv:  true
	// Send:  CT_SPAWN
	// Recv:  true
	// Send:  CT_WAIT
	// Recv:  true
}
