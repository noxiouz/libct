package libct

import "bytes"
import "io"
import "testing"
import "syscall"
import "os"

func TestCreateCT(t *testing.T) {
	stdin, err := os.OpenFile("/dev/null", os.O_RDWR, 0);
	r, w, err := os.Pipe()
	pipes := Pipes {int(stdin.Fd()), int(w.Fd()), int(stdin.Fd())}

	s, err := OpenSession()
	if err != nil {
		t.Fatal(err)
	}

	ct, err := s.CreateCt("test")
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
	err = ct.CtExecve("/bin/bash", argv, env, &pipes)
	if err != nil {
		t.Fatal(err)
	}

	w.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r)
	r.Close()
	t.Log(buf);

	// wait
	err = ct.CtWait()
	if err != nil {
		t.Fatal(err)
	}
}
