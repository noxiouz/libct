package libct

import "net"
import "fmt"
import "syscall"
import prot "code.google.com/p/goprotobuf/proto"

type Session struct {
	sk *net.UnixConn
}

type LibctError struct {
	Code int32
}

func (e LibctError) Error() string {
	return fmt.Sprintf("LibctError: %x", e.Code)
}

func OpenSession() (*Session, error) {
	addr, err := net.ResolveUnixAddr("unixpacket", "/var/run/libct.sock")
	if err != nil {
		return nil, err
	}
	sk, err := net.DialUnix("unixpacket", nil, addr)
	if err != nil {
		return nil, err
	}

	return &Session{sk}, nil
}

type Container struct {
	s   *Session
	Rid uint64
}

func __sendReq(s *Session, req *RpcRequest) (error) {

	fmt.Println("Send: ", req.Req.String())

	pkt, err := prot.Marshal(req)
	if err != nil {
		return err
	}

	s.sk.Write(pkt)
	return nil
}

func __recvRes(s *Session) (*RpcResponce, error) {

	pkt := make([]byte, 4096)
	size, err := s.sk.Read(pkt)
	if err != nil {
		return nil, err
	}

	res := &RpcResponce{}
	err = prot.Unmarshal(pkt[0:size], res)
	if err != nil {
		return nil, err
	}

	fmt.Println("Recv: ", res.GetSuccess())
	if !res.GetSuccess() {
		return nil, LibctError{res.GetError()}
	}

	return res, nil
}

func sendReq(s *Session, req *RpcRequest) (*RpcResponce, error) {
	err := __sendReq(s, req)
	if err != nil {
		return nil, err
	}

	return __recvRes(s)
}

func (s *Session) CreateCt(name string) (*Container, error) {
	req := &RpcRequest{}

	req.Req = ReqType_CT_CREATE.Enum()

	req.Create = &CreateReq{
		Name: prot.String(name),
	}

	res, err := sendReq(s, req)
	if err != nil {
		return nil, err
	}

	return &Container{s, res.Create.GetRid()}, nil
}

type Pipes struct {
	stdin, stdout, stderr int;
}

func (ct *Container) CtExecve(path string, argv []string, env []string, pipes *Pipes) error {
	pipes_here := (pipes != nil)
	req := &RpcRequest{}

	req.Req = ReqType_CT_SPAWN.Enum()
	req.CtRid = &ct.Rid

	req.Execv = &ExecvReq{
		Path: &path,
		Args: argv,
		Env:  env,
		Pipes: &pipes_here,
	}

	err := __sendReq(ct.s, req)
	if err != nil {
		return err
	}

	if pipes_here {
		rights := syscall.UnixRights(pipes.stdin, pipes.stdout, pipes.stderr)
		dummyByte := []byte("x")
		_, _, err = ct.s.sk.WriteMsgUnix(dummyByte, rights, nil)
		if err != nil {
			return err
		}
	}

	_, err = __recvRes(ct.s)

	return err
}

func (ct *Container) CtWait() error {
	req := &RpcRequest{}

	req.Req = ReqType_CT_WAIT.Enum()
	req.CtRid = &ct.Rid

	_, err := sendReq(ct.s, req)

	return err
}

func (ct *Container) SetNsMask(nsmask uint64) error {
	req := &RpcRequest{}
	req.Req = ReqType_CT_SETNSMASK.Enum()
	req.CtRid = &ct.Rid
	req.Nsmask = &NsmaskReq{Mask : &nsmask}

	_, err := sendReq(ct.s, req)

	return err
}

func (ct *Container)SetFsRoot(root string) error {
	req := &RpcRequest{}
	req.Req = ReqType_FS_SETROOT.Enum()
	req.CtRid = &ct.Rid
	req.Setroot = &SetrootReq{Root : &root}

	_, err := sendReq(ct.s, req)

	return err
}

const (
	CT_FS_NONE	= 0
	CT_FS_SUBDIR	= 1
)

func (ct *Container)SetFsPrivate(ptype int32, path string) error {
	req := &RpcRequest{}
	req.Req = ReqType_FS_SETPRIVATE.Enum()
	req.CtRid = &ct.Rid
	req.Setpriv = &SetprivReq{Type : &ptype, Path : &path}

	_, err := sendReq(ct.s, req)

	return err
}
