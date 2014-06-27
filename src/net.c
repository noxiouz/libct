#include <stdio.h>
#include <sched.h>

#include <netinet/ether.h>

#include <netlink/netlink.h>
#include <netlink/route/link.h>
#include <netlink/route/link/veth.h>

#include "uapi/libct.h"

#include "xmalloc.h"
#include "list.h"
#include "err.h"
#include "net.h"
#include "ct.h"

/*
 * Generic Linux networking management
 */

/*
 * Move network device @name into task's @pid net namespace
 */

static struct nl_sock *net_get_sock()
{
	static struct nl_sock *sk;
	int err;

	if (sk)
		return sk;

	sk = nl_socket_alloc();
        if ((err = nl_connect(sk, NETLINK_ROUTE)) < 0) {
                nl_perror(err, "Unable to connect socket");
                return NULL;
        }

	return sk;
}

static int net_nic_chnage(char *name, struct rtnl_link *link)
{
	struct rtnl_link *orig;
	struct nl_sock *sk;
	int err = -1;

	sk = net_get_sock();
	if (sk == NULL)
		return -1;

	orig = rtnl_link_alloc();

	if (!orig)
		goto free;

	rtnl_link_set_name(orig, name);

	if ((err = rtnl_link_change(sk, orig, link, 0)) < 0) {
                nl_perror(err, "Unable to change link");
                return err;
        }

free:
	rtnl_link_put(orig);
	return err;
}

/*
 * VETH creation/removal
 */

#ifndef VETH_INFO_MAX
enum {
	VETH_INFO_UNSPEC,
	VETH_INFO_PEER,

	__VETH_INFO_MAX
#define VETH_INFO_MAX   (__VETH_INFO_MAX - 1)
};
#endif

struct ct_net_veth {
	struct ct_net n;
};

static int veth_pair_create(struct rtnl_link *link, int ct_pid)
{
	struct nl_sock *sk;
	int err;

	sk = net_get_sock();
	if (sk == NULL)
		return -1;

	rtnl_link_set_ns_pid(link, ct_pid);

	if ((err = rtnl_link_add(sk, link, NLM_F_CREATE)) < 0) {
                nl_perror(err, "Unable to add link");
                return err;
        }

	return 0;
}

/*
 * Library API implementation
 */

void net_release(struct container *ct)
{
	struct ct_net *cn, *n;

	list_for_each_entry_safe(cn, n, &ct->ct_nets, l) {
		list_del(&cn->l);
		cn->ops->destroy(cn);
	}
}

int net_start(struct container *ct)
{
	struct ct_net *cn;

	list_for_each_entry(cn, &ct->ct_nets, l) {
		if (cn->ops->start(ct, cn))
			goto err;
	}

	return 0;

err:
	list_for_each_entry_continue_reverse(cn, &ct->ct_nets, l)
		cn->ops->stop(ct, cn);
	return -1;
}

void net_stop(struct container *ct)
{
	struct ct_net *cn;

	list_for_each_entry(cn, &ct->ct_nets, l)
		cn->ops->stop(ct, cn);
}

net_dev_t local_net_add(ct_handler_t h, enum ct_net_type ntype, void *arg)
{
	struct container *ct = cth2ct(h);
	const struct ct_net_ops *nops;
	struct ct_net *cn;

	if (ct->state != CT_STOPPED)
		/* FIXME -- implement */
		return ERR_PTR(-LCTERR_BADCTSTATE);

	if (!(ct->nsmask & CLONE_NEWNET))
		return ERR_PTR(-LCTERR_NONS);

	if (ntype == CT_NET_NONE)
		return 0;

	nops = net_get_ops(ntype);
	if (!nops)
		return ERR_PTR(-LCTERR_BADTYPE);

	cn = nops->create(arg);
	if (!cn)
		return ERR_PTR(-LCTERR_BADARG);

	cn->ops = nops;
	list_add_tail(&cn->l, &ct->ct_nets);
	return (net_dev_t) cn->link;
}

int local_net_del(ct_handler_t h, enum ct_net_type ntype, void *arg)
{
	struct container *ct = cth2ct(h);
	const struct ct_net_ops *nops;
	struct ct_net *cn;

	if (ct->state != CT_STOPPED)
		/* FIXME -- implement */
		return -LCTERR_BADCTSTATE;

	if (ntype == CT_NET_NONE)
		return 0;

	nops = net_get_ops(ntype);
	if (!nops)
		return -LCTERR_BADTYPE;

	list_for_each_entry(cn, &ct->ct_nets, l) {
		if (!cn->ops->match(cn, arg))
			continue;

		list_del(&cn->l);
		cn->ops->destroy(cn);
		return 0;
	}

	return -LCTERR_NOTFOUND;
}

net_dev_t libct_net_add(ct_handler_t ct, enum ct_net_type ntype, void *arg)
{
	return ct->ops->net_add(ct, ntype, arg);
}

int libct_net_del(ct_handler_t ct, enum ct_net_type ntype, void *arg)
{
	return ct->ops->net_del(ct, ntype, arg);
}

/*
 * CT_NET_HOSTNIC management
 */

struct ct_net_host_nic {
	struct ct_net n;
	char *name;
};

static inline struct ct_net_host_nic *cn2hn(struct ct_net *n)
{
	return container_of(n, struct ct_net_host_nic, n);
}

static struct ct_net *host_nic_create(void *arg)
{
	struct ct_net_host_nic *cn;

	if (arg) {
		cn = xmalloc(sizeof(*cn));
		if (cn) {
			cn->n.link = rtnl_link_alloc();
			cn->name = xstrdup(arg);
			if (cn->name && cn->n.link) {
				rtnl_link_set_name(cn->n.link, cn->name);
				return &cn->n;
			}
			xfree(cn->name);
			rtnl_link_put(cn->n.link);
		}
		xfree(cn);
	}
	return NULL;
}

static void host_nic_destroy(struct ct_net *n)
{
	struct ct_net_host_nic *cn = cn2hn(n);

	xfree(cn->name);
	xfree(cn);
}

static int host_nic_start(struct container *ct, struct ct_net *n)
{
	rtnl_link_set_ns_pid(n->link, ct->root_pid);
	return net_nic_chnage(cn2hn(n)->name, n->link);
}

static void host_nic_stop(struct container *ct, struct ct_net *n)
{
	/* 
	 * Nothing to do here. On container stop it's NICs will
	 * just jump out of it.
	 *
	 * FIXME -- CT owner might have changed NIC name. Handle
	 * it by checking the NIC's index.
	 */
}

static int host_nic_match(struct ct_net *n, void *arg)
{
	struct ct_net_host_nic *cn = cn2hn(n);
	return !strcmp(cn->name, arg);
}

static const struct ct_net_ops host_nic_ops = {
	.create		= host_nic_create,
	.destroy	= host_nic_destroy,
	.start		= host_nic_start,
	.stop		= host_nic_stop,
	.match		= host_nic_match,
};

/*
 * CT_NET_VETH management
 */

static struct ct_net_veth *cn2vn(struct ct_net *n)
{
	return container_of(n, struct ct_net_veth, n);
}

static void veth_free(struct ct_net_veth *vn)
{
	rtnl_link_veth_release(vn->n.link);
	xfree(vn);
}

static struct ct_net *veth_create(void *arg)
{
	struct ct_net_veth_arg *va = arg;
	struct ct_net_veth *vn;
	struct rtnl_link *peer;

	if (!arg || !va->host_name || !va->ct_name)
		return NULL;

	vn = xmalloc(sizeof(*vn));
	if (!vn)
		return NULL;

	vn->n.link = rtnl_link_veth_alloc();
	if (!vn->n.link) {
		veth_free(vn);
		return NULL;
	}

	rtnl_link_set_name(vn->n.link, va->ct_name);
        peer = rtnl_link_veth_get_peer(vn->n.link);
        rtnl_link_set_name(peer, va->host_name);
	rtnl_link_put(peer);

	return &vn->n;
}

static void veth_destroy(struct ct_net *n)
{
	veth_free(cn2vn(n));
}

static int veth_start(struct container *ct, struct ct_net *n)
{
	if (veth_pair_create(n->link, ct->root_pid))
		return -1;

	return 0;
}

static void veth_stop(struct container *ct, struct ct_net *n)
{
	/* 
	 * FIXME -- don't destroy veth here, keep it across
	 * container's restarts. This needs checks in the
	 * veth_pair_create() for existance.
	 */
}

static int veth_match(struct ct_net *n, void *arg)
{
	struct ct_net_veth_arg *va = arg;
	struct rtnl_link *peer;
	char *host_name;

	peer = rtnl_link_veth_get_peer(n->link);
	host_name = rtnl_link_get_name(peer);

	/* Matching hostname should be enough */
	return !strcmp(host_name, va->host_name);
}

static const struct ct_net_ops veth_nic_ops = {
	.create		= veth_create,
	.destroy	= veth_destroy,
	.start		= veth_start,
	.stop		= veth_stop,
	.match		= veth_match,
};

const struct ct_net_ops *net_get_ops(enum ct_net_type ntype)
{
	switch (ntype) {
	case CT_NET_HOSTNIC:
		return &host_nic_ops;
	case CT_NET_VETH:
		return &veth_nic_ops;
	case CT_NET_NONE:
		break;
	}

	return NULL;
}

int libct_net_dev_set_mac(net_dev_t d, char *mac)
{
	struct rtnl_link *link = (struct rtnl_link *) d;
	struct nl_addr* addr;

	addr = nl_addr_build(AF_LLC, ether_aton(mac), ETH_ALEN);
	rtnl_link_set_addr(link, addr);

	return 0;
}
