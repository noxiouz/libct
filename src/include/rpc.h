#ifndef __LIBCT_RPC_H__
#define __LIBCT_RPC_H__

#include "uapi/libct.h"

struct _RpcRequest;
typedef struct _RpcRequest RpcRequest;

enum {
	CT_STATE,
};

typedef int (rpc_callback)(libct_session_t s, RpcRequest *req, void *req_args, int type, void *args);

int rpc_async_add(RpcRequest *req, void *args);
int rpc_async_run(libct_session_t s, int type, void *args);

#endif
