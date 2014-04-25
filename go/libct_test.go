package libct

import "testing"

func TestCreateCT(t *testing.T) {
	s, err := OpenSession()
	if err != nil {
		t.Fatal(err)
	}

	ct, err := s.CreateCt()
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}

	// wait
	err = ct.CtWait()
	if err != nil {
		t.Fatal(err)
	}
}
