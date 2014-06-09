package libct

// #cgo CFLAGS: -DCONFIG_X86_64 -DARCH="x86" -D_FILE_OFFSET_BITS=64 -D_GNU_SOURCE
// #cgo LDFLAGS: -lct
// #include "../src/include/uapi/libct.h"
// #include "../src/include/uapi/libct-errors.h"
import "C"
import "fmt"

type Session struct {
	s C.libct_session_t
}

type Container struct {
	ct C.ct_handler_t
}

type LibctError struct {
	Code int32
}

func (e LibctError) Error() string {
	return fmt.Sprintf("LibctError: %x", e.Code)
}

func (s *Session) OpenLocal() error {
	s.s = C.libct_session_open_local()

	if s.s == nil {
		return LibctError{-1}
	}

	return nil
}

func (s *Session) ContainerCreate(name string) (*Container, error) {
	ct := C.libct_container_create(s.s, C.CString(name))

	if ct == nil {
		return nil, LibctError{-1}
	}

	return &Container{ct}, nil
}

func (ct *Container) SetNsMask(nsmask uint64) error {
	ret := C.libct_container_set_nsmask(ct.ct, C.ulong(nsmask))

	if int(ret) != 0 {
		return LibctError{int32(ret)}
	}

	return nil
}

func (ct *Container) SpawnExecve(path string, argv []string, env []string) error {
	cargv := make([]*C.char, len(argv)+1)
	for i, arg := range argv {
		cargv[i] = C.CString(arg)
	}

	cenv := make([]*C.char, len(env)+1)
	for i, e := range env {
		cenv[i] = C.CString(e)
	}

	ret := C.libct_container_spawn_execve(ct.ct, C.CString(path), &cargv[0], &cenv[0])
	if int(ret) != 0 {
		return LibctError{int32(ret)}
	}

	return nil
}

func (ct *Container) Wait() error {
	ret := C.libct_container_wait(ct.ct)

	if int(ret) != 0 {
		return LibctError{int32(ret)}
	}

	return nil
}
