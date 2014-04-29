// Code generated by protoc-gen-go.
// source: rpc.proto
// DO NOT EDIT!

package libct

import proto "code.google.com/p/goprotobuf/proto"
import json "encoding/json"
import math "math"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

type ReqType int32

const (
	ReqType_EMPTY         ReqType = 0
	ReqType_CT_CREATE     ReqType = 1
	ReqType_CT_OPEN       ReqType = 2
	ReqType_CT_DESTROY    ReqType = 10
	ReqType_CT_GET_STATE  ReqType = 11
	ReqType_CT_SPAWN      ReqType = 12
	ReqType_CT_ENTER      ReqType = 13
	ReqType_CT_KILL       ReqType = 14
	ReqType_CT_WAIT       ReqType = 15
	ReqType_CT_SETNSMASK  ReqType = 16
	ReqType_CT_ADD_CNTL   ReqType = 17
	ReqType_CT_CFG_CNTL   ReqType = 18
	ReqType_FS_SETROOT    ReqType = 19
	ReqType_FS_SETPRIVATE ReqType = 20
	ReqType_FS_ADD_MOUNT  ReqType = 21
	ReqType_FS_DEL_MOUNT  ReqType = 22
	ReqType_CT_SET_OPTION ReqType = 23
	ReqType_CT_NET_ADD    ReqType = 24
	ReqType_CT_NET_DEL    ReqType = 25
	ReqType_CT_UNAME      ReqType = 26
	ReqType_CT_SET_CAPS   ReqType = 27
)

var ReqType_name = map[int32]string{
	0:  "EMPTY",
	1:  "CT_CREATE",
	2:  "CT_OPEN",
	10: "CT_DESTROY",
	11: "CT_GET_STATE",
	12: "CT_SPAWN",
	13: "CT_ENTER",
	14: "CT_KILL",
	15: "CT_WAIT",
	16: "CT_SETNSMASK",
	17: "CT_ADD_CNTL",
	18: "CT_CFG_CNTL",
	19: "FS_SETROOT",
	20: "FS_SETPRIVATE",
	21: "FS_ADD_MOUNT",
	22: "FS_DEL_MOUNT",
	23: "CT_SET_OPTION",
	24: "CT_NET_ADD",
	25: "CT_NET_DEL",
	26: "CT_UNAME",
	27: "CT_SET_CAPS",
}
var ReqType_value = map[string]int32{
	"EMPTY":         0,
	"CT_CREATE":     1,
	"CT_OPEN":       2,
	"CT_DESTROY":    10,
	"CT_GET_STATE":  11,
	"CT_SPAWN":      12,
	"CT_ENTER":      13,
	"CT_KILL":       14,
	"CT_WAIT":       15,
	"CT_SETNSMASK":  16,
	"CT_ADD_CNTL":   17,
	"CT_CFG_CNTL":   18,
	"FS_SETROOT":    19,
	"FS_SETPRIVATE": 20,
	"FS_ADD_MOUNT":  21,
	"FS_DEL_MOUNT":  22,
	"CT_SET_OPTION": 23,
	"CT_NET_ADD":    24,
	"CT_NET_DEL":    25,
	"CT_UNAME":      26,
	"CT_SET_CAPS":   27,
}

func (x ReqType) Enum() *ReqType {
	p := new(ReqType)
	*p = x
	return p
}
func (x ReqType) String() string {
	return proto.EnumName(ReqType_name, int32(x))
}
func (x ReqType) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.String())
}
func (x *ReqType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(ReqType_value, data, "ReqType")
	if err != nil {
		return err
	}
	*x = ReqType(value)
	return nil
}

type CreateReq struct {
	Name             *string `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *CreateReq) Reset()         { *m = CreateReq{} }
func (m *CreateReq) String() string { return proto.CompactTextString(m) }
func (*CreateReq) ProtoMessage()    {}

func (m *CreateReq) GetName() string {
	if m != nil && m.Name != nil {
		return *m.Name
	}
	return ""
}

type ExecvReq struct {
	Path             *string  `protobuf:"bytes,1,req,name=path" json:"path,omitempty"`
	Args             []string `protobuf:"bytes,2,rep,name=args" json:"args,omitempty"`
	Env              []string `protobuf:"bytes,3,rep,name=env" json:"env,omitempty"`
	Pipes            *bool    `protobuf:"varint,4,opt,name=pipes,def=0" json:"pipes,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *ExecvReq) Reset()         { *m = ExecvReq{} }
func (m *ExecvReq) String() string { return proto.CompactTextString(m) }
func (*ExecvReq) ProtoMessage()    {}

const Default_ExecvReq_Pipes bool = false

func (m *ExecvReq) GetPath() string {
	if m != nil && m.Path != nil {
		return *m.Path
	}
	return ""
}

func (m *ExecvReq) GetArgs() []string {
	if m != nil {
		return m.Args
	}
	return nil
}

func (m *ExecvReq) GetEnv() []string {
	if m != nil {
		return m.Env
	}
	return nil
}

func (m *ExecvReq) GetPipes() bool {
	if m != nil && m.Pipes != nil {
		return *m.Pipes
	}
	return Default_ExecvReq_Pipes
}

type NsmaskReq struct {
	Mask             *uint64 `protobuf:"varint,1,req,name=mask" json:"mask,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *NsmaskReq) Reset()         { *m = NsmaskReq{} }
func (m *NsmaskReq) String() string { return proto.CompactTextString(m) }
func (*NsmaskReq) ProtoMessage()    {}

func (m *NsmaskReq) GetMask() uint64 {
	if m != nil && m.Mask != nil {
		return *m.Mask
	}
	return 0
}

type AddcntlReq struct {
	Ctype            *uint32 `protobuf:"varint,1,req,name=ctype" json:"ctype,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *AddcntlReq) Reset()         { *m = AddcntlReq{} }
func (m *AddcntlReq) String() string { return proto.CompactTextString(m) }
func (*AddcntlReq) ProtoMessage()    {}

func (m *AddcntlReq) GetCtype() uint32 {
	if m != nil && m.Ctype != nil {
		return *m.Ctype
	}
	return 0
}

type CfgcntlReq struct {
	Ctype            *uint32 `protobuf:"varint,1,req,name=ctype" json:"ctype,omitempty"`
	Param            *string `protobuf:"bytes,2,req,name=param" json:"param,omitempty"`
	Value            *string `protobuf:"bytes,3,req,name=value" json:"value,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *CfgcntlReq) Reset()         { *m = CfgcntlReq{} }
func (m *CfgcntlReq) String() string { return proto.CompactTextString(m) }
func (*CfgcntlReq) ProtoMessage()    {}

func (m *CfgcntlReq) GetCtype() uint32 {
	if m != nil && m.Ctype != nil {
		return *m.Ctype
	}
	return 0
}

func (m *CfgcntlReq) GetParam() string {
	if m != nil && m.Param != nil {
		return *m.Param
	}
	return ""
}

func (m *CfgcntlReq) GetValue() string {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return ""
}

type SetrootReq struct {
	Root             *string `protobuf:"bytes,1,req,name=root" json:"root,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *SetrootReq) Reset()         { *m = SetrootReq{} }
func (m *SetrootReq) String() string { return proto.CompactTextString(m) }
func (*SetrootReq) ProtoMessage()    {}

func (m *SetrootReq) GetRoot() string {
	if m != nil && m.Root != nil {
		return *m.Root
	}
	return ""
}

type SetprivReq struct {
	Type             *int32  `protobuf:"varint,1,req,name=type" json:"type,omitempty"`
	Path             *string `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *SetprivReq) Reset()         { *m = SetprivReq{} }
func (m *SetprivReq) String() string { return proto.CompactTextString(m) }
func (*SetprivReq) ProtoMessage()    {}

func (m *SetprivReq) GetType() int32 {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return 0
}

func (m *SetprivReq) GetPath() string {
	if m != nil && m.Path != nil {
		return *m.Path
	}
	return ""
}

type SetoptionReq struct {
	Opt              *int32  `protobuf:"varint,1,req,name=opt" json:"opt,omitempty"`
	CgPath           *string `protobuf:"bytes,2,opt,name=cg_path" json:"cg_path,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *SetoptionReq) Reset()         { *m = SetoptionReq{} }
func (m *SetoptionReq) String() string { return proto.CompactTextString(m) }
func (*SetoptionReq) ProtoMessage()    {}

func (m *SetoptionReq) GetOpt() int32 {
	if m != nil && m.Opt != nil {
		return *m.Opt
	}
	return 0
}

func (m *SetoptionReq) GetCgPath() string {
	if m != nil && m.CgPath != nil {
		return *m.CgPath
	}
	return ""
}

type NetaddReq struct {
	Type             *int32  `protobuf:"varint,1,req,name=type" json:"type,omitempty"`
	Nicname          *string `protobuf:"bytes,2,opt,name=nicname" json:"nicname,omitempty"`
	Peername         *string `protobuf:"bytes,3,opt,name=peername" json:"peername,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *NetaddReq) Reset()         { *m = NetaddReq{} }
func (m *NetaddReq) String() string { return proto.CompactTextString(m) }
func (*NetaddReq) ProtoMessage()    {}

func (m *NetaddReq) GetType() int32 {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return 0
}

func (m *NetaddReq) GetNicname() string {
	if m != nil && m.Nicname != nil {
		return *m.Nicname
	}
	return ""
}

func (m *NetaddReq) GetPeername() string {
	if m != nil && m.Peername != nil {
		return *m.Peername
	}
	return ""
}

type MountReq struct {
	Dst              *string `protobuf:"bytes,1,req,name=dst" json:"dst,omitempty"`
	Src              *string `protobuf:"bytes,2,opt,name=src" json:"src,omitempty"`
	Flags            *int32  `protobuf:"varint,3,opt,name=flags" json:"flags,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *MountReq) Reset()         { *m = MountReq{} }
func (m *MountReq) String() string { return proto.CompactTextString(m) }
func (*MountReq) ProtoMessage()    {}

func (m *MountReq) GetDst() string {
	if m != nil && m.Dst != nil {
		return *m.Dst
	}
	return ""
}

func (m *MountReq) GetSrc() string {
	if m != nil && m.Src != nil {
		return *m.Src
	}
	return ""
}

func (m *MountReq) GetFlags() int32 {
	if m != nil && m.Flags != nil {
		return *m.Flags
	}
	return 0
}

type UnameReq struct {
	Host             *string `protobuf:"bytes,1,opt,name=host" json:"host,omitempty"`
	Domain           *string `protobuf:"bytes,2,opt,name=domain" json:"domain,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *UnameReq) Reset()         { *m = UnameReq{} }
func (m *UnameReq) String() string { return proto.CompactTextString(m) }
func (*UnameReq) ProtoMessage()    {}

func (m *UnameReq) GetHost() string {
	if m != nil && m.Host != nil {
		return *m.Host
	}
	return ""
}

func (m *UnameReq) GetDomain() string {
	if m != nil && m.Domain != nil {
		return *m.Domain
	}
	return ""
}

type CapsReq struct {
	ApplyTo          *uint32 `protobuf:"varint,1,req,name=apply_to" json:"apply_to,omitempty"`
	Mask             *uint64 `protobuf:"varint,2,req,name=mask" json:"mask,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *CapsReq) Reset()         { *m = CapsReq{} }
func (m *CapsReq) String() string { return proto.CompactTextString(m) }
func (*CapsReq) ProtoMessage()    {}

func (m *CapsReq) GetApplyTo() uint32 {
	if m != nil && m.ApplyTo != nil {
		return *m.ApplyTo
	}
	return 0
}

func (m *CapsReq) GetMask() uint64 {
	if m != nil && m.Mask != nil {
		return *m.Mask
	}
	return 0
}

type CreateResp struct {
	Rid              *uint64 `protobuf:"varint,1,req,name=rid" json:"rid,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *CreateResp) Reset()         { *m = CreateResp{} }
func (m *CreateResp) String() string { return proto.CompactTextString(m) }
func (*CreateResp) ProtoMessage()    {}

func (m *CreateResp) GetRid() uint64 {
	if m != nil && m.Rid != nil {
		return *m.Rid
	}
	return 0
}

type StateResp struct {
	State            *uint32 `protobuf:"varint,1,req,name=state" json:"state,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *StateResp) Reset()         { *m = StateResp{} }
func (m *StateResp) String() string { return proto.CompactTextString(m) }
func (*StateResp) ProtoMessage()    {}

func (m *StateResp) GetState() uint32 {
	if m != nil && m.State != nil {
		return *m.State
	}
	return 0
}

type ExecvResp struct {
	Pid              *int32 `protobuf:"varint,1,req,name=pid" json:"pid,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *ExecvResp) Reset()         { *m = ExecvResp{} }
func (m *ExecvResp) String() string { return proto.CompactTextString(m) }
func (*ExecvResp) ProtoMessage()    {}

func (m *ExecvResp) GetPid() int32 {
	if m != nil && m.Pid != nil {
		return *m.Pid
	}
	return 0
}

type RpcRequest struct {
	Req              *ReqType      `protobuf:"varint,1,req,name=req,enum=ReqType" json:"req,omitempty"`
	CtRid            *uint64       `protobuf:"varint,2,opt,name=ct_rid" json:"ct_rid,omitempty"`
	Create           *CreateReq    `protobuf:"bytes,3,opt,name=create" json:"create,omitempty"`
	Execv            *ExecvReq     `protobuf:"bytes,4,opt,name=execv" json:"execv,omitempty"`
	Nsmask           *NsmaskReq    `protobuf:"bytes,5,opt,name=nsmask" json:"nsmask,omitempty"`
	Addcntl          *AddcntlReq   `protobuf:"bytes,6,opt,name=addcntl" json:"addcntl,omitempty"`
	Cfgcntl          *CfgcntlReq   `protobuf:"bytes,7,opt,name=cfgcntl" json:"cfgcntl,omitempty"`
	Setroot          *SetrootReq   `protobuf:"bytes,8,opt,name=setroot" json:"setroot,omitempty"`
	Setpriv          *SetprivReq   `protobuf:"bytes,9,opt,name=setpriv" json:"setpriv,omitempty"`
	Setopt           *SetoptionReq `protobuf:"bytes,10,opt,name=setopt" json:"setopt,omitempty"`
	Netadd           *NetaddReq    `protobuf:"bytes,11,opt,name=netadd" json:"netadd,omitempty"`
	Mnt              *MountReq     `protobuf:"bytes,12,opt,name=mnt" json:"mnt,omitempty"`
	Uname            *UnameReq     `protobuf:"bytes,13,opt,name=uname" json:"uname,omitempty"`
	Caps             *CapsReq      `protobuf:"bytes,14,opt,name=caps" json:"caps,omitempty"`
	XXX_unrecognized []byte        `json:"-"`
}

func (m *RpcRequest) Reset()         { *m = RpcRequest{} }
func (m *RpcRequest) String() string { return proto.CompactTextString(m) }
func (*RpcRequest) ProtoMessage()    {}

func (m *RpcRequest) GetReq() ReqType {
	if m != nil && m.Req != nil {
		return *m.Req
	}
	return 0
}

func (m *RpcRequest) GetCtRid() uint64 {
	if m != nil && m.CtRid != nil {
		return *m.CtRid
	}
	return 0
}

func (m *RpcRequest) GetCreate() *CreateReq {
	if m != nil {
		return m.Create
	}
	return nil
}

func (m *RpcRequest) GetExecv() *ExecvReq {
	if m != nil {
		return m.Execv
	}
	return nil
}

func (m *RpcRequest) GetNsmask() *NsmaskReq {
	if m != nil {
		return m.Nsmask
	}
	return nil
}

func (m *RpcRequest) GetAddcntl() *AddcntlReq {
	if m != nil {
		return m.Addcntl
	}
	return nil
}

func (m *RpcRequest) GetCfgcntl() *CfgcntlReq {
	if m != nil {
		return m.Cfgcntl
	}
	return nil
}

func (m *RpcRequest) GetSetroot() *SetrootReq {
	if m != nil {
		return m.Setroot
	}
	return nil
}

func (m *RpcRequest) GetSetpriv() *SetprivReq {
	if m != nil {
		return m.Setpriv
	}
	return nil
}

func (m *RpcRequest) GetSetopt() *SetoptionReq {
	if m != nil {
		return m.Setopt
	}
	return nil
}

func (m *RpcRequest) GetNetadd() *NetaddReq {
	if m != nil {
		return m.Netadd
	}
	return nil
}

func (m *RpcRequest) GetMnt() *MountReq {
	if m != nil {
		return m.Mnt
	}
	return nil
}

func (m *RpcRequest) GetUname() *UnameReq {
	if m != nil {
		return m.Uname
	}
	return nil
}

func (m *RpcRequest) GetCaps() *CapsReq {
	if m != nil {
		return m.Caps
	}
	return nil
}

type RpcResponce struct {
	Success          *bool       `protobuf:"varint,1,req,name=success" json:"success,omitempty"`
	Error            *int32      `protobuf:"varint,2,opt,name=error" json:"error,omitempty"`
	Create           *CreateResp `protobuf:"bytes,3,opt,name=create" json:"create,omitempty"`
	State            *StateResp  `protobuf:"bytes,4,opt,name=state" json:"state,omitempty"`
	Execv            *ExecvResp  `protobuf:"bytes,5,opt,name=execv" json:"execv,omitempty"`
	XXX_unrecognized []byte      `json:"-"`
}

func (m *RpcResponce) Reset()         { *m = RpcResponce{} }
func (m *RpcResponce) String() string { return proto.CompactTextString(m) }
func (*RpcResponce) ProtoMessage()    {}

func (m *RpcResponce) GetSuccess() bool {
	if m != nil && m.Success != nil {
		return *m.Success
	}
	return false
}

func (m *RpcResponce) GetError() int32 {
	if m != nil && m.Error != nil {
		return *m.Error
	}
	return 0
}

func (m *RpcResponce) GetCreate() *CreateResp {
	if m != nil {
		return m.Create
	}
	return nil
}

func (m *RpcResponce) GetState() *StateResp {
	if m != nil {
		return m.State
	}
	return nil
}

func (m *RpcResponce) GetExecv() *ExecvResp {
	if m != nil {
		return m.Execv
	}
	return nil
}

func init() {
	proto.RegisterEnum("ReqType", ReqType_name, ReqType_value)
}
