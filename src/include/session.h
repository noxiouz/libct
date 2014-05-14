#ifndef __LIBCT_SESSION_H__
#define __LIBCT_SESSION_H__

#include "list.h"
#include "ct.h"
#include "rpc.h"

struct container;

enum {
	BACKEND_NONE,
	BACKEND_LOCAL,
	BACKEND_UNIX,
};

struct backend_ops {
	int type;
	ct_handler_t (*create_ct)(libct_session_t s, char *name);
	ct_handler_t (*open_ct)(libct_session_t s, char *name);
	void	     (*update_ct_state)(libct_session_t s, pid_t pid);
	void (*close)(libct_session_t s);
	rpc_callback *async_cb;
};

struct libct_session {
	struct backend_ops *ops;
	struct list_head s_cts;
};

struct local_session {
	struct libct_session s;
	int server_sk;
};

static inline struct local_session *s2ls(libct_session_t s)
{
	return container_of(s, struct local_session, s);
}

void local_session_add(libct_session_t, struct container *);
void update_container_state(libct_session_t s, pid_t pid);

#endif /* __LIBCT_SESSION_H__ */
