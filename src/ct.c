#include <sched.h>
#include <signal.h>
#include <unistd.h>
#include <fcntl.h>
#include <grp.h>
#include <stdarg.h>
#include <stdlib.h>
#include <limits.h>
#include <errno.h>

#include <sys/wait.h>
#include <sys/mount.h>
#include <sys/prctl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <dirent.h>
#include <sys/socket.h>

#include "uapi/libct.h"
#include "asm/page.h"

#include "linux-kernel.h"
#include "namespaces.h"
#include "xmalloc.h"
#include "session.h"
#include "cgroups.h"
#include "security.h"
#include "list.h"
#include "util.h"
#include "lsm.h"
#include "net.h"
#include "err.h"
#include "ct.h"
#include "fs.h"
#include "vz.h"

static enum ct_state local_get_state(ct_handler_t h)
{
	return cth2ct(h)->state;
}

static void local_ct_uid_gid_free(struct container *ct)
{
	struct _uid_gid_map *map, *t;

	list_for_each_entry_safe(map, t, &ct->uid_map, node)
		xfree(map);
	list_for_each_entry_safe(map, t, &ct->gid_map, node)
		xfree(map);
}

static void local_ct_destroy(ct_handler_t h)
{
	struct container *ct = cth2ct(h);

	cgroups_free(ct);
	fs_free(ct);
	net_release(ct);
	xfree(ct->name);
	xfree(ct->hostname);
	xfree(ct->domainname);
	xfree(ct->cgroup_sub);
	local_ct_uid_gid_free(ct);
	xfree(ct);
}

static int local_set_nsmask(ct_handler_t h, unsigned long nsmask)
{
	struct container *ct = cth2ct(h);

	if (ct->state != CT_STOPPED)
		return -LCTERR_BADCTSTATE;

	/* Are all of these bits supported by kernel? */
	if (nsmask & ~kernel_ns_mask)
		return -LCTERR_NONS;

	if (!(nsmask & CLONE_NEWNS))
		net_release(ct);
	if (!(nsmask & CLONE_NEWUTS)) {
		xfree(ct->hostname);
		ct->hostname = NULL;
		xfree(ct->domainname);
		ct->domainname = NULL;
	}

	ct->nsmask = nsmask;
	return 0;
}

struct ct_clone_arg {
	char stack[PAGE_SIZE] __attribute__((aligned (8)));
	char stack_ptr[0];
	int (*cb)(void *);
	void *arg;
	struct container *ct;
	struct process_desc *p;
	int wait_sock[2];
	bool is_exec;
};

static int re_mount_proc(struct container *ct)
{
	if (!ct->root_path) {
		if (mount("none", "/proc", "none", MS_PRIVATE|MS_REC, NULL))
			return -1;

		umount2("/proc", MNT_DETACH);
	}

	return mount("proc", "/proc", "proc", 0, NULL);
}

static int try_mount_proc(struct container *ct)
{
	/* Not requested by user */
	if (!(ct->flags & CT_AUTO_PROC))
		return 0;

	/* Container w/o pidns can work on existing proc */
	if (!(ct->nsmask & CLONE_NEWPID) && !ct->root_path)
		return 0;

	/* Container with shared FS has no place for new proc */
	if (!fs_private(ct))
		return -1;

	return re_mount_proc(ct);
}

extern int pivot_root(const char *new_root, const char *put_old);

static int set_current_root(char *path)
{
	if (chroot(path))
		return -1;
	if (chdir("/"))
		return -1;
	return 0;
}

static int set_ct_root(struct container *ct)
{
	char put_root[] = "libct-root.XXXXXX";

	if (!(ct->nsmask & CLONE_NEWNS))
		return set_current_root(ct->root_path);

	/*
	 * We're in new mount namespace. No need in
	 * just going into chroot, do pivot root, that
	 * gives us the ability to umount old tree.
	 */

	if (mount(ct->root_path, ct->root_path, NULL, MS_BIND | MS_REC, NULL) == -1)
		return -1;

	if (chdir(ct->root_path))
		return -1;

	if (mkdtemp(put_root) == NULL)
		return -1;

	if (pivot_root(".", put_root)) {
		rmdir(put_root);
		return -1;
	}

	if (umount2(put_root, MNT_DETACH))
		return -1;

	rmdir(put_root);
	return 0;
}

static int uname_set(struct container *ct)
{
	int ret = 0;

	if (ct->hostname)
		ret |= sethostname(ct->hostname, strlen(ct->hostname));

	if (ct->domainname)
		ret |= setdomainname(ct->domainname, strlen(ct->domainname));

	return ret;
}

static int apply_env(struct process_desc *p)
{
	int i;

	if (clearenv()) {
		pr_perror("Unable to clear the environment");
	}

	for (i = 0; i < p->envn; i++) {
		char *c;

		c = strchr(p->env[i], '=');
		*c = 0;
		if (setenv(p->env[i], c + 1, 1)) {
			*c = '=';
			pr_perror("Unable to set %s", p->env[i]);
			return -1;
		}
		*c = '=';
	}
	return 0;
}

static int ct_clone(void *arg)
{
	int ret = -1;
	struct ct_clone_arg *ca = arg;
	struct container *ct = ca->ct;
	struct process_desc *p = ca->p;
	int wait_sock = ca->wait_sock[1];

	close(ca->wait_sock[0]);

	ret = spawn_sock_wait_and_close(wait_sock);
	if (ret)
		goto err_um;

	if (ct->nsmask & CLONE_NEWUSER) {
		if (setuid(0) || setgid(0) || setgroups(0, NULL))
			goto err;
	}

	if (prctl(PR_SET_PDEATHSIG, p->pdeathsig))
		goto err;

	if (!(ct->flags & CT_NOSETSID) && setsid() == -1)
		goto err;

	if (ct->tty_fd == LIBCT_CONSOLE_FD) {
		ct->tty_fd = open("/dev/console", O_RDWR);
		if (ct->tty_fd == -1) {
			pr_perror("Unable to open /dev/console");
			return -1;
		}
	}

	if (ct->tty_fd >= 0 && ioctl(ct->tty_fd, TIOCSCTTY, 0) == -1)
		goto err;

	if (p->fds) {
		p->fds[p->fdn] = wait_sock;

		if (setup_fds(p->fds, p->fdn + 1))
			goto err;

		wait_sock = p->fdn;
		if (fcntl(wait_sock, F_SETFD, FD_CLOEXEC)) {
			goto err;
		}
	}

	if (ct->nsmask & CLONE_NEWNS) {
		/*
		 * Remount / as slave, so that it doesn't
		 * propagate its changes to our container.
		 */
		ret = -LCTERR_CANTMOUNT;
		if (mount("none", "/", "none", MS_SLAVE|MS_REC, NULL))
			goto err;
	}

	if (try_mount_cg(ct))
		goto err;

	if (ct->root_path) {
		/*
		 * Mount external in child, since it may live
		 * in sub mount namespace. If it doesn't do
		 * it here anyway, just umount by hands in the
		 * fs_umount().
		 */
		ret = fs_mount_ext(ct);
		if (ret < 0)
			goto err;

		ret = set_ct_root(ct);
		if (ret < 0)
			goto err_um;

		ret = fs_create_devnodes(ct);
		if (ret < 0)
			goto err;
	}

	ret = uname_set(ct);
	if (ret < 0)
		goto err_um;

	ret = try_mount_proc(ct);
	if (ret < 0)
		goto err_um;

	ret = apply_creds(p);
	if (ret < 0)
		goto err_um;

	ret = apply_env(p);
	if (ret < 0)
		goto err_um;

	if (p->lsm_label)
		ret = lsm_process_label_set(p->lsm_label, false, p->lsm_on_exec);
	p->lsm_on_exec = 0;
	if (ret < 0)
		goto err;

	spawn_sock_wake(wait_sock, 0);
	if (!ca->is_exec)
		close(wait_sock);

	ret = ca->cb(ca->arg);
	if (ca->is_exec)
		goto err;

	return ret;

err_um:
	if (ct->root_path)
		fs_umount_ext(ct);
err:
	if (ret >= 0)
		ret = -1;
	spawn_sock_wake(wait_sock, ret);
	close(wait_sock);
	exit(ret);
}

static int write_id_mappings(pid_t pid, struct list_head *list, char *id_map)
{
	int size = 0, off = 0, exit_code, fd = -1;
	struct _uid_gid_map *map;
	char *buf = NULL, *_buf;
	char fname[PATH_MAX];

	list_for_each_entry(map, list, node) {
		if (size - off < 34) {
			size += PAGE_SIZE;
			_buf = xrealloc(buf, size);
			if (_buf == NULL)
				goto err;
			buf = _buf;
		}
		off += snprintf(buf + off, size - off, "%u %u %u\n",
				map->first, map->lower_first, map->count);

	}

	snprintf(fname, sizeof(fname), "/proc/%d/%s", pid, id_map);
	fd = open(fname, O_WRONLY);
	if (fd < 0)
		goto err;
	if (write(fd, buf, off) != off)
		goto err;

	exit_code = 0;
err:
	xfree(buf);
	if (fd > 0)
		close(fd);
	return exit_code;
}

static ct_process_t __local_spawn_cb(ct_handler_t h, ct_process_desc_t ph, int (*cb)(void *), void *arg, bool is_exec)
{
	struct container *ct = cth2ct(h);
	struct process_desc *p = prh2pr(ph);
	int ret = -1, pid, aux;
	struct ct_clone_arg ca;
	int wait_sock;

	if (ct->state != CT_STOPPED)
		return ERR_PTR(-LCTERR_BADCTSTATE);

	ret = fs_mount(ct);
	if (ret)
		return ERR_PTR(ret);

	if ((ct->flags & CT_KILLABLE) && !(ct->nsmask & CLONE_NEWPID)) {
		if (add_service_controller(ct))
			goto err_cg;
	}

	ret = cgroups_create(ct);
	if (ret)
		goto err_cg;

	ret = -1;
	if (socketpair(AF_FILE, SOCK_SEQPACKET | SOCK_CLOEXEC, 0, ca.wait_sock)) {
		pr_perror("Unable to create a socket pair");
		goto err_sock;
	}
	wait_sock = ca.wait_sock[0];

	ca.cb = cb;
	ca.arg = arg;
	ca.ct = ct;
	ca.p = p;
	ca.is_exec = is_exec;
	pid = clone(ct_clone, &ca.stack_ptr, ct->nsmask | SIGCHLD, &ca);
	close(ca.wait_sock[1]);
	if (pid < 0)
		goto err_clone;

	ct->p.pid = pid;

	if (ct->nsmask & CLONE_NEWUSER) {
		if (write_id_mappings(pid, &ct->uid_map, "uid_map"))
			goto err_net;

		if (write_id_mappings(pid, &ct->gid_map, "gid_map"))
			goto err_net;
	}

	ret = cgroups_attach(ct, pid);
	if (ret < 0)
		goto err_net;

	if (net_start(ct))
		goto err_net;

	spawn_sock_wake_and_close(wait_sock, 0);

	aux = spawn_sock_wait(wait_sock);
	if (aux != 0) {
		ret = aux;
		goto err_ch;
	}

	aux = spawn_sock_wait(wait_sock);
	if (aux != INT_MIN) {
		ret = -1;
		goto err_ch;
	}
	close(wait_sock);

	ct->state = CT_RUNNING;
	return &ct->p.h;

err_ch:
	net_stop(ct);
err_net:
	spawn_sock_wake_and_close(wait_sock, -1);
	libct_process_wait(&ct->p.h, NULL);
err_clone:
	close(wait_sock);
err_sock:
	cgroups_destroy(ct);
err_cg:
	fs_umount(ct);
	return ERR_PTR(ret);
}

static ct_process_t local_spawn_cb(ct_handler_t h, ct_process_desc_t ph, int (*cb)(void *), void *arg)
{
	return __local_spawn_cb(h, ph, cb, arg, false);
}

static int ct_execv(void *a)
{
	struct execv_args *ea = a;
	sigset_t mask;

	sigfillset(&mask);
	sigprocmask(SIG_UNBLOCK, &mask, NULL);

	/* This gets control in the container's new root (if any) */
	if (ea->env)
		execvpe(ea->path, ea->argv, ea->env);
	else
		execvp(ea->path, ea->argv);

	return -1;
}

static ct_process_t local_spawn_execve(ct_handler_t ct, ct_process_desc_t pr, char *path, char **argv, char **env)
{
	struct execv_args ea;
	struct process_desc *p = prh2pr(pr);

	ea.path = path;
	ea.argv = argv;
	ea.env = env;

	p->lsm_on_exec = true;

	return __local_spawn_cb(ct, pr, ct_execv, &ea, true);
}

static ct_process_t __local_enter_cb(ct_handler_t h, ct_process_desc_t ph, int (*cb)(void *), void *arg, bool is_exec)
{
	struct container *ct = cth2ct(h);
	struct process_desc *p = prh2pr(ph);
	struct process *pr;
	int aux = -1, pid = -1;
	int wait_socks[2];
	int wait_sock = -1;

	if (ct->state != CT_RUNNING)
		return ERR_PTR(-LCTERR_BADCTSTATE);

	if (ct->nsmask & CLONE_NEWPID) {
		if (switch_ns(ct->p.pid, &pid_ns, &aux))
			return ERR_PTR(-LCTERR_INVARG);
	}

	pr = xmalloc(sizeof(struct process));
	if (pr == NULL)
		return ERR_PTR(-1);

	local_process_init(pr);

	if (socketpair(AF_FILE, SOCK_SEQPACKET | SOCK_CLOEXEC, 0, wait_socks)) {
		pr_perror("Unable to create a socket pair");
		goto err;
	}
	wait_sock = wait_socks[0];

	pid = fork();
	if (pid == 0) {
		struct ns_desc *ns;
		wait_sock = wait_socks[1];

		close(wait_socks[0]);

		if (spawn_sock_wait_and_close(wait_sock)) /* wait cgroups */
			exit(1);

		if (p->fds) {
			p->fds[p->fdn] = wait_sock;

			if (setup_fds(p->fds, p->fdn + 1))
				exit(-1);

			wait_sock = p->fdn;
			if (fcntl(wait_sock, F_SETFD, FD_CLOEXEC)) {
				goto err;
			}
		}

		for (aux = 0; namespaces[aux]; aux++) {
			ns = namespaces[aux];

			if (ns->cflag == CLONE_NEWPID)
				continue;
			if (!(ns->cflag & ct->nsmask))
				continue;

			if (switch_ns(ct->p.pid, ns, NULL))
				exit(-1);
		}

		if (ct->root_path && !(ct->nsmask & CLONE_NEWNS)) {
			char nroot[128];

			/*
			 * Otherwise switched by setns()
			 */

			snprintf(nroot, sizeof(nroot), "/proc/%d/root", ct->p.pid);
			if (set_current_root(nroot))
				exit(-1);
		}

		if (apply_creds(p))
			exit(-1);

		if (apply_env(p))
			exit(-1);

		spawn_sock_wake(wait_sock, 0);
		if (!is_exec)
			close(wait_sock);

		aux = cb(arg);

		if (is_exec)
			spawn_sock_wake_and_close(wait_sock, -1);
		exit(aux);
	}
	close(wait_socks[1]);
	if (pid < 0)
		goto err;

	if (aux >= 0)
		restore_ns(aux, &pid_ns);

	if (cgroups_attach(ct, pid))
		goto err;

	spawn_sock_wake_and_close(wait_sock, 0);

	if (spawn_sock_wait(wait_sock))
		goto err;

	if (spawn_sock_wait_and_close(wait_sock) != INT_MIN)
		goto err;

	pr->pid = pid;

	return &pr->h;
err:
	xfree(pr);
	close(wait_sock);
	if (pid > 0)
		waitpid(pid, NULL, 0);
	return ERR_PTR(-1);
}

static ct_process_t local_enter_cb(ct_handler_t h, ct_process_desc_t ph, int (*cb)(void *), void *arg)
{
	return __local_enter_cb(h, ph, cb, arg, false);
}

static ct_process_t local_enter_execve(ct_handler_t h, ct_process_desc_t pr, char *path, char **argv, char **env)
{
	struct execv_args ea = {};
	struct process_desc *p = prh2pr(pr);

	ea.path	= path;
	ea.argv	= argv;
	ea.env	= env;

	p->lsm_on_exec = true;

	return __local_enter_cb(h, pr, ct_execv, &ea, true);
}

static int local_ct_kill(ct_handler_t h)
{
	struct container *ct = cth2ct(h);

	if (ct->state != CT_RUNNING)
		return -LCTERR_BADCTSTATE;
	if (ct->nsmask & CLONE_NEWPID)
		return kill(ct->p.pid, SIGKILL);
	if (ct->flags & CT_KILLABLE)
		return service_ctl_killall(ct);
	return -1;
}

static int local_ct_wait(ct_handler_t h)
{
	struct container *ct = cth2ct(h);
	int ret, status;

	if (ct->state != CT_RUNNING)
		return -LCTERR_BADCTSTATE;

	if (ct->p.pid > 0) {
		ret = libct_process_wait(&ct->p.h, &status);
		if (ret < 0)
			return -1;
	}

	fs_umount(ct);
	cgroups_destroy(ct); /* FIXME -- can be held accross restarts */
	net_stop(ct);

	ct->state = CT_STOPPED;
	return 0;
}

static int local_set_option(ct_handler_t h, int opt, void *args)
{
	int ret = -LCTERR_BADTYPE;
	struct container *ct = cth2ct(h);

	switch (opt) {
	case LIBCT_OPT_AUTO_PROC_MOUNT:
		ret = 0;
		ct->flags |= CT_AUTO_PROC;
		break;
	case LIBCT_OPT_CGROUP_SUBMOUNT:
		ret = 0;
		if (args)
			ct->cgroup_sub = xstrdup((char *) args);
		else
			ct->cgroup_sub = xstrdup(DEFAULT_CGROUPS_PATH);
		if (!ct->cgroup_sub)
			ret = -1;
		break;
	case LIBCT_OPT_KILLABLE:
		ret = cgroups_create_service();
		if (!ret)
			ct->flags |= CT_KILLABLE;
		break;
	case LIBCT_OPT_NOSETSID:
		ret = 0;
		ct->flags |= CT_NOSETSID;
		break;
	}

	return ret;
}

static int local_uname(ct_handler_t h, char *host, char *dom)
{
	struct container *ct = cth2ct(h);

	if (!(ct->nsmask & CLONE_NEWUTS))
		return -LCTERR_NONS;
	if (ct->state != CT_STOPPED)
		return -LCTERR_BADCTSTATE; /* FIXME */

	if (host) {
		host = xstrdup(host);
		if (!host)
			return -1;
	}
	xfree(ct->hostname);
	ct->hostname = host;

	if (dom) {
		dom = xstrdup(dom);
		if (!dom)
			return -1;
	}
	xfree(ct->domainname);
	ct->domainname = dom;

	return 0;
}

char *local_ct_name(ct_handler_t h)
{
	return cth2ct(h)->name;
}

static int local_set_console_fd(ct_handler_t h, int fd)
{
	struct container *ct = cth2ct(h);
	ct->tty_fd = fd;
	return 0;
}

static int local_add_map(struct list_head *list, unsigned int first,
			unsigned int lower_first, unsigned int count)
{
	struct _uid_gid_map *_map;

	_map = xmalloc(sizeof(struct _uid_gid_map));
	if (_map == NULL)
		return -1;

	_map->first		= first;
	_map->lower_first	= lower_first;
	_map->count		= count;

	list_add(&_map->node, list);

	return 0;
}

static int local_add_uid_map(ct_handler_t h, unsigned int first,
			unsigned int lower_first, unsigned int count)
{
	struct container *ct = cth2ct(h);

	return local_add_map(&ct->uid_map, first, lower_first, count);
}

static int local_add_gid_map(ct_handler_t h, unsigned int first,
			unsigned int lower_first, unsigned int count)
{
	struct container *ct = cth2ct(h);

	return local_add_map(&ct->gid_map, first, lower_first, count);
}

static const struct container_ops local_ct_ops = {
	.spawn_cb		= local_spawn_cb,
	.spawn_execve		= local_spawn_execve,
	.enter_cb		= local_enter_cb,
	.enter_execve		= local_enter_execve,
	.kill			= local_ct_kill,
	.wait			= local_ct_wait,
	.destroy		= local_ct_destroy,
	.detach			= local_ct_destroy,
	.set_nsmask		= local_set_nsmask,
	.add_controller		= local_add_controller,
	.config_controller	= local_config_controller,
	.fs_set_root		= local_fs_set_root,
	.fs_set_private		= local_fs_set_private,
	.fs_add_mount		= local_add_mount,
	.fs_add_bind_mount	= local_add_bind_mount,
	.fs_del_bind_mount	= local_del_bind_mount,
	.fs_add_devnode		= local_add_devnode,
	.get_state		= local_get_state,
	.set_option		= local_set_option,
	.set_console_fd		= local_set_console_fd,
	.net_add		= local_net_add,
	.net_del		= local_net_del,
	.net_route_add		= local_net_route_add,
	.uname			= local_uname,
	.add_uid_map		= local_add_uid_map,
	.add_gid_map		= local_add_gid_map,
	.get_processes		= local_controller_tasks,
};

ct_handler_t ct_create(char *name)
{
	struct container *ct;

	ct = xzalloc(sizeof(*ct));
	if (ct) {
		ct_handler_init(&ct->h);
		ct->h.ops = &local_ct_ops;
		ct->state = CT_STOPPED;
		ct->name = xstrdup(name);
		ct->tty_fd = -1;
		INIT_LIST_HEAD(&ct->cgroups);
		INIT_LIST_HEAD(&ct->cg_configs);
		INIT_LIST_HEAD(&ct->ct_nets);
		INIT_LIST_HEAD(&ct->ct_net_routes);
		INIT_LIST_HEAD(&ct->fs_mnts);
		INIT_LIST_HEAD(&ct->fs_devnodes);
		INIT_LIST_HEAD(&ct->uid_map);
		INIT_LIST_HEAD(&ct->gid_map);

		local_process_init(&ct->p);

		return &ct->h;
	}

	return NULL;
}
