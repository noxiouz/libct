package libct

// #cgo CFLAGS: -DCONFIG_X86_64 -DARCH="x86" -D_FILE_OFFSET_BITS=64 -D_GNU_SOURCE
// #cgo LDFLAGS: -lct
// #include "../src/include/uapi/libct.h"
// #include "../src/include/uapi/libct-errors.h"
import "C"
import "fmt"
import "os"

const (
	LIBCT_OPT_AUTO_PROC_MOUNT = C.LIBCT_OPT_AUTO_PROC_MOUNT
)

type Session struct {
	s C.libct_session_t
}

type Container struct {
	ct C.ct_handler_t
}

type LibctError struct {
	Code int
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

	if ret != 0 {
		return LibctError{int(ret)}
	}

	return nil
}

func (ct *Container) SetConsoleFd(f *os.File) error {
	ret := C.libct_container_set_console_fd(ct.ct, C.int(f.Fd()))

	if ret != 0 {
		return LibctError{int(ret)}
	}

	return nil
}

func (ct *Container) SpawnExecve(path string, argv []string, env []string, fds *[3]uintptr) error {
	var cfdsp *C.int

	cargv := make([]*C.char, len(argv)+1)
	for i, arg := range argv {
		cargv[i] = C.CString(arg)
	}

	cenv := make([]*C.char, len(env)+1)
	for i, e := range env {
		cenv[i] = C.CString(e)
	}

	if fds != nil {
		cfds := make([]C.int, 3)
		for i, fd := range fds {
			cfds[i] = C.int(fd)
		}
		cfdsp = &cfds[0]
	}

	ret := C.libct_container_spawn_execvefds(ct.ct, C.CString(path), &cargv[0], &cenv[0], cfdsp)
	if ret != 0 {
		return LibctError{int(ret)}
	}

	return nil
}

func (ct *Container) Wait() error {
	ret := C.libct_container_wait(ct.ct)

	if ret != 0 {
		return LibctError{int(ret)}
	}

	return nil
}

func (ct *Container) Uname(host *string, domain *string) error {
	var chost *C.char
	var cdomain *C.char

	if host != nil {
		chost = C.CString(*host)
	}

	if domain != nil {
		cdomain = C.CString(*domain)
	}

	ret := C.libct_container_uname(ct.ct, chost, cdomain)

	if ret != 0 {
		return LibctError{int(ret)}
	}

	return nil
}

func (ct *Container) SetRoot(root string) error {

	if ret := C.libct_fs_set_root(ct.ct, C.CString(root)); ret != 0 {
		return LibctError{int(ret)}
	}

	return nil
}

const (
	CT_FS_RDONLY  = C.CT_FS_RDONLY
	CT_FS_PRIVATE = C.CT_FS_PRIVATE
)

func (ct *Container) AddMount(src string, dst string, flags int) error {

	if ret := C.libct_fs_add_mount(ct.ct, C.CString(src), C.CString(dst), C.int(flags)); ret != 0 {
		return LibctError{int(ret)}
	}

	return nil
}

func (ct *Container) SetOption(opt int32) error {
	if ret := C.libct_container_set_option(ct.ct, C.int(opt), nil); ret != 0 {
		return LibctError{int(ret)}
	}

	return nil
}

func (ct *Container) AddDeviceNode(path string, mode int, major int, minor int) error {

	ret := C.libct_fs_add_devnode(ct.ct, C.CString(path), C.int(mode), C.int(major), C.int(minor))

	if ret != 0 {
		return LibctError{int(ret)}
	}

	return nil
}
