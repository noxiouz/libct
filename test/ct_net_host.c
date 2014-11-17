/*
 * Test how host nic assignment works
 */
#define _GNU_SOURCE
#include <sys/types.h>
#include <libct.h>
#include <stdio.h>
#include <sys/mman.h>
#include <sched.h>
#include <stdlib.h>

#include "test.h"

static int check_ct_net(void *a)
{
	int *ct_status = a;

	ct_status[0] = 1;
	if (!system("ip link l dm0"))
		ct_status[1] = 1;

	return 0;
}

int main(int argc, char **argv)
{
	int *ct_status;
	libct_session_t s;
	ct_handler_t ct;
	ct_process_desc_t p;
	ct_net_t nd;

	ct_status = mmap(NULL, 4096, PROT_READ | PROT_WRITE,
			MAP_SHARED | MAP_ANON, 0, 0);
	ct_status[0] = 0;
	ct_status[1] = 0;

	system("ip link add name dm0 type dummy");
	if (system("ip link l dm0"))
		return err("Can't create dummy device");

	s = libct_session_open_local();
	ct = libct_container_create(s, "test");
	p = libct_process_desc_create(s);
	libct_container_set_nsmask(ct, CLONE_NEWNET);

	nd = libct_net_add(ct, CT_NET_HOSTNIC, "dm0");
	if (libct_handle_is_err(nd)) {
		system("ip link del dm0");
		return err("Can't add hostnic");
	}
	if (libct_container_spawn_cb(ct, p, check_ct_net, ct_status) < 0) {
		system("ip link del dm0");
		return err("Can't spawn CT");
	}

	libct_container_wait(ct);
	libct_container_destroy(ct);
	libct_session_close(s);

	system("ip link del dm0");

	if (!ct_status[0])
		return fail("CT is not alive");
	if (!ct_status[1])
		return fail("Netdevice not assigned");

	return pass("HostNic works OK");
}
