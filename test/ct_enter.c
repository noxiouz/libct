/*
 * Test entering into living container
 */
#include <libct.h>
#include <stdio.h>
#include <sys/mman.h>
#include <sys/wait.h>
#include <unistd.h>

#include "test.h"

struct ct_arg {
	int wait_fd;
	int *mark;
};

static int set_ct_alive(void *a)
{
	struct ct_arg *cta = a;
	char c;

	cta->mark[0] = 1;
	read(cta->wait_fd, &c, 1);
	return 0;
}

static int set_ct_enter(void *a)
{
	struct ct_arg *cta = a;
	cta->mark[1] = 1;
	return 0;
}

int main(int argc, char **argv)
{
	struct ct_arg cta;
	int pid, p[2];
	libct_session_t s;
	ct_handler_t ct;

	pipe(p);
	cta.mark = mmap(NULL, 4096, PROT_READ | PROT_WRITE,
			MAP_SHARED | MAP_ANON, 0, 0);
	cta.mark[0] = 0;
	cta.mark[1] = 0;
	cta.wait_fd = p[0];

	s = libct_session_open_local();
	ct = libct_container_create(s, "test");
	libct_container_spawn_cb(ct, set_ct_alive, &cta);
	pid = libct_container_enter_cb(ct, set_ct_enter, &cta);
	waitpid(pid, NULL, 0);
	write(p[1], "a", 1);
	libct_container_wait(ct);
	libct_container_destroy(ct);
	libct_session_close(s);

	if (!cta.mark[0])
		return fail("CT is not alive");

	if (!cta.mark[1])
		return fail("CT is not enterable");

	return pass("CT is created and entered");
}
