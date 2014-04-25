package libct

import "testing"
import "syscall"

func TestCreateCT(t *testing.T) {
	s, err := OpenSession()
	if err != nil {
		t.Fatal(err)
	}

	ct, err := s.CreateCt()
	if err != nil {
		t.Fatal(err)
	}

	err = ct.SetNsMask(syscall.CLONE_NEWNS | syscall.CLONE_NEWPID)
	if err != nil {
		t.Fatal(err)
	}

	err = ct.SetFsPrivate(CT_FS_SUBDIR, "/home/avagin/centos")
	if err != nil {
		t.Fatal(err)
	}
	err = ct.SetFsRoot("/home/avagin/centos-root")
	if err != nil {
		t.Fatal(err)
	}

	// exec
	argv := make([]string, 3)
	argv[0] = "bash"
	argv[1] = "-c"
	argv[2] = "echo Hello"
//	argv[2] = "sleep 10"
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
