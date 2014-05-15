package libct

import "net"
import "fmt"
import "syscall"
import prot "code.google.com/p/goprotobuf/proto"

type Session struct {
	sk *net.UnixConn
	resp_map map[uint64]chan *RpcResponce
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

	s := &Session{sk, map[uint64]chan *RpcResponce{}}
	go func() {
		for {
			resp, err := __recvRes(s)
			if err != nil {
				return
			}
			fmt.Println("Send in chanel", *resp.ReqId)
			s.resp_map[*resp.ReqId] <- resp
			fmt.Println("Send in chanel done")
		}
	}()

	return s, nil
}

func (s *Session)_sendReq(req *RpcRequest, pipes *Pipes) (chan *RpcResponce, error) {
	c := make(chan *RpcResponce)
	s.resp_map[*req.ReqId] = c

	err := __sendReq(s, req, pipes)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type Container struct {
	s   *Session
	Rid uint64
	pid int32
}

func __sendReq(s *Session, req *RpcRequest, pipes *Pipes) (error) {

	fmt.Println("Send: ", req.Req.String(), *req.ReqId)

	pkt, err := prot.Marshal(req)
	if err != nil {
		return err
	}

	rights := syscall.UnixRights() //FIXME
	if pipes != nil {
		rights = syscall.UnixRights(pipes.Stdin, pipes.Stdout, pipes.Stderr)
	}

	_, _, err = s.sk.WriteMsgUnix(pkt, rights, nil)
	if err != nil {
		return err
	}
	fmt.Println("Send: done", req.Req.String())
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

func sendReqWithPipes(s *Session, req *RpcRequest, pipes *Pipes) (*RpcResponce, error) {
	c, err := s._sendReq(req, pipes)
	if err != nil {
		return nil, err
	}

	fmt.Println("Wait resp on ", req.Req.String())
	resp := <-c
	return resp, nil
}

func sendReq(s *Session, req *RpcRequest) (*RpcResponce, error) {
	return sendReqWithPipes(s, req, nil)
}

var id uint64 = 0;

func getRpcReq() (*RpcRequest) {
	req := &RpcRequest{}
	i := id
	id++;
	fmt.Println("getRpcReq ", id)
	req.ReqId = &i;
	return req
}

func (s *Session) CreateCt(name string) (*Container, error) {
	req := getRpcReq()

	req.Req = ReqType_CT_CREATE.Enum()

	req.Create = &CreateReq{
		Name: prot.String(name),
	}

	res, err := sendReq(s, req)
	if err != nil {
		return nil, err
	}

	return &Container{s, res.Create.GetRid(), 0}, nil
}

func (s *Session) OpenCt(name string) (*Container, error) {
	req := getRpcReq()

	req.Req = ReqType_CT_OPEN.Enum()

	req.Create = &CreateReq{
		Name: prot.String(name),
	}

	res, err := sendReq(s, req)
	if err != nil {
		return nil, err
	}

	return &Container{s, res.Create.GetRid(), 0}, nil
}

type Pipes struct {
	Stdin, Stdout, Stderr int;
}

func (ct *Container) Run(path string, argv []string, env []string, pipes *Pipes) (int32, error) {
	pipes_here := (pipes != nil)
	req := getRpcReq()

	req.Req = ReqType_CT_SPAWN.Enum()
	req.CtRid = &ct.Rid

	req.Execv = &ExecvReq{
		Path: &path,
		Args: argv,
		Env:  env,
		Pipes: &pipes_here,
	}

	resp, err := sendReqWithPipes(ct.s, req, pipes)
	if err != nil {
		return 0, err
	}


	return resp.Execv.GetPid(), err
}

func (ct *Container) Wait() error {
	req := getRpcReq()

	req.Req = ReqType_CT_WAIT.Enum()
	req.CtRid = &ct.Rid

	_, err := sendReq(ct.s, req)

	return err
}

func (ct *Container) Kill() error {
	req := getRpcReq()

	req.Req = ReqType_CT_KILL.Enum()
	req.CtRid = &ct.Rid

	_, err := sendReq(ct.s, req)

	return err
}

const (
	CT_ERROR int	= -1
	CT_STOPPED	= 0
	CT_RUNNING	= 1
)

func (ct *Container) State() (int, error) {
	req := getRpcReq()

	req.Req = ReqType_CT_GET_STATE.Enum()
	req.CtRid = &ct.Rid

	resp, err := sendReq(ct.s, req)
	if err != nil {
		return CT_ERROR, err
	}

	return int(resp.State.GetState()), nil
}

func (ct *Container) SetNsMask(nsmask uint64) error {
	req := getRpcReq()
	req.Req = ReqType_CT_SETNSMASK.Enum()
	req.CtRid = &ct.Rid
	req.Nsmask = &NsmaskReq{Mask : &nsmask}

	_, err := sendReq(ct.s, req)

	return err
}

func (ct *Container)SetFsRoot(root string) error {
	req := getRpcReq()
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
	req := getRpcReq()
	req.Req = ReqType_FS_SETPRIVATE.Enum()
	req.CtRid = &ct.Rid
	req.Setpriv = &SetprivReq{Type : &ptype, Path : &path}

	_, err := sendReq(ct.s, req)

	return err
}

func (ct *Container)AddMount(src, dst string) error {
	req := getRpcReq()
	req.Req = ReqType_FS_ADD_MOUNT.Enum()
	req.CtRid = &ct.Rid
	flags := int32(0)
	req.Mnt = &MountReq{
				Dst : &dst,
				Src : &src,
				Flags : &flags,
			}

	_, err := sendReq(ct.s, req)

	return err
}

func (ct *Container)SetOption(opt int32) error {
	req := getRpcReq()
	req.Req = ReqType_CT_SET_OPTION.Enum()
	req.CtRid = &ct.Rid
	req.Setopt = &SetoptionReq{ Opt : &opt}

	_, err := sendReq(ct.s, req)

	return err
}
