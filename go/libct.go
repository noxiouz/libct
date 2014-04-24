package libct

import "net"
import "fmt"
import prot "code.google.com/p/goprotobuf/proto"

type Session struct {
	sk net.Conn
}

type LibctError struct {
	Code int32
}

func (e LibctError) Error() string {
	return fmt.Sprintf("LibctError: %x", e.Error)
}

func OpenSession() (*Session, error) {
	sk, err := net.Dial("unixpacket", "/var/run/libct.sock")
	if err != nil {
		return nil, err
	}

	return &Session{sk}, nil
}

type Container struct {
	s   *Session
	Rid uint64
}

func sendReq(s *Session, req *RpcRequest) (*RpcResponce, error) {

	fmt.Println("Send: ", req.Req.String())

	pkt, err := prot.Marshal(req)
	if err != nil {
		return nil, err
	}

	s.sk.Write(pkt)

	pkt = make([]byte, 4096)
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

func (s *Session) CreateCt() (*Container, error) {
	req := &RpcRequest{}

	req.Req = ReqType_CT_CREATE.Enum()

	req.Create = &CreateReq{
		Name: prot.String("test"),
	}

	res, err := sendReq(s, req)
	if err != nil {
		return nil, err
	}

	return &Container{s, res.Create.GetRid()}, nil
}

func (ct *Container) CtExecve(path string, argv []string, env []string) error {
	req := &RpcRequest{}

	req.Req = ReqType_CT_SPAWN.Enum()
	req.CtRid = &ct.Rid

	req.Execv = &ExecvReq{
		Path: &path,
		Args: argv,
		Env:  env,
	}

	_, err := sendReq(ct.s, req)

	return err
}

func (ct *Container) CtWait() error {
	req := &RpcRequest{}

	req.Req = ReqType_CT_WAIT.Enum()
	req.CtRid = &ct.Rid

	_, err := sendReq(ct.s, req)

	return err
}
