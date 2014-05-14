#include <stdlib.h>

#include "list.h"
#include "xmalloc.h"
#include "rpc.h"
#include "session.h"

#include "protobuf/rpc.pb-c.h"

struct rpc_async_req {
	struct list_head node;
	RpcRequest *req;
	void *args;
};

static LIST_HEAD(rpc_async_list); // FIXME libct_session_t

int rpc_async_add(RpcRequest *req, void *args)
{
	struct rpc_async_req *r;

	r = xmalloc(sizeof(struct rpc_async_req));
	if (r == NULL)
		return -1;

	r->req = req;
	r->args = args;
	list_add(&r->node, &rpc_async_list);

	return 1;
}

int rpc_async_run(libct_session_t s, int type, void *args)
{
	struct rpc_async_req *req;
	int ret;

	list_for_each_entry(req, &rpc_async_list, node) {
		ret = s->ops->async_cb(s, req->req, req->args, type,  args);
		if (ret < 0)
			return -1;
		if (ret == 1)
			list_del(&req->node);
	}

	return 0;
}
