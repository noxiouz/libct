/*
 * Test empty "container" creation
 */
#include <libct.h>
#include <stdio.h>
#include <sys/mman.h>
#include <linux/sched.h>
#include "test.h"

static int set_ct_alive(void *a)
{
	if (getuid() != 0)
		return -1;
	if (getgid() != 0)
		return -1;
	*(int *)a = 1;
	return 0;
}

int main(int argc, char **argv)
{
	int *ct_alive;
	libct_session_t s;
	ct_handler_t ct;

	ct_alive = mmap(NULL, 4096, PROT_READ | PROT_WRITE,
			MAP_SHARED | MAP_ANON, 0, 0);
	*ct_alive = 0;

	s = libct_session_open_local();
	ct = libct_container_create(s, "test");
	if (libct_container_set_nsmask(ct, CLONE_NEWPID | CLONE_NEWUSER))
		return 1;
	if (libct_userns_add_uid_map(ct, 0, 120000, 1100) ||
	    libct_userns_add_uid_map(ct, 1100, 130000, 1200) ||
	    libct_userns_add_gid_map(ct, 0, 140000, 1200) ||
	    libct_userns_add_gid_map(ct, 1200, 150000, 1100))
		return 1;
	libct_container_spawn_cb(ct, set_ct_alive, ct_alive);
	libct_container_wait(ct);
	libct_container_destroy(ct);
	libct_session_close(s);

	if (!*ct_alive)
		return fail("Container is not alive");
	else
		return pass("Container is alive");
}
