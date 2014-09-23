/*
 * Test that veth pair can be created
 */
#define _GNU_SOURCE
#include <sys/types.h>
#include <libct.h>
#include <stdio.h>
#include <sys/mman.h>
#include <sched.h>
#include <unistd.h>
#include "test.h"

#define VETH_HOST_NAME	"hveth0"
#define VETH_CT_NAME	"cveth0"

struct ct_arg {
	int wait_pipe;
	int *mark;
};

static int check_ct_net(void *a)
{
	struct ct_arg *ca = a;
	char c;

	ca->mark[0] = 1;
	if (!system("ip a l " VETH_CT_NAME ""))
		ca->mark[2] = 1;

	system("ip r");

	read(ca->wait_pipe, &c, 1);
	return 0;
}

int main(int argc, char **argv)
{
	int p[2];
	struct ct_arg ca;
	libct_session_t s;
	ct_handler_t ct;
	struct ct_net_veth_arg va;
	ct_net_t nd, peer;
	ct_net_route_t r;
	ct_net_route_nh_t nh;

	ca.mark = mmap(NULL, 4096, PROT_READ | PROT_WRITE,
			MAP_SHARED | MAP_ANON, 0, 0);
	pipe(p);

	ca.mark[0] = 0;
	ca.mark[1] = 0;
	ca.mark[2] = 0;
	ca.wait_pipe = p[0];

	va.host_name = VETH_HOST_NAME;
	va.ct_name = VETH_CT_NAME;

	s = libct_session_open_local();
	ct = libct_container_create(s, "test");
	libct_container_set_nsmask(ct, CLONE_NEWNET);

	nd = libct_net_add(ct, CT_NET_VETH, &va);
	if (libct_handle_is_err(nd))
		return err("Can't add hostnic");

	if (libct_net_dev_set_mac_addr(nd, "00:11:22:33:44:55"))
		return err("Can't set mac");

	if (libct_net_dev_add_ip_addr(nd, "192.168.87.124/24"))
		return err("Can't set ip");

	peer = libct_net_dev_get_peer(nd);

	r = libct_net_route_add(ct);
	if (r == NULL)
		return err("Can't allocate an route entry");

	libct_net_route_set_dst(r, "192.168.88.0/24");
	libct_net_route_set_dev(r, "cveth0");

	nh = libct_net_route_add_nh(r);
	libct_net_route_nh_set_gw(nh, "192.168.87.123/24");

	if (libct_container_spawn_cb(ct, check_ct_net, &ca) < 0)
		return err("Can't spawn CT");

	if (!system("ip a l " VETH_HOST_NAME ""))
		ca.mark[1] = 1;

	write(p[1], "a", 1);

	libct_container_wait(ct);
	libct_container_destroy(ct);
	libct_session_close(s);

	if (!ca.mark[0])
		return fail("CT is not alive");
	if (!ca.mark[1])
		return fail("VETH not created");
	if (!ca.mark[2])
		return fail("VETH not assigned");

	return pass("VETH works OK");
}
