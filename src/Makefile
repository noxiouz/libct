lib-so			+= libct

obj-y			+= libct.o
obj-y			+= session.o
obj-y			+= ct.o
obj-y			+= namespaces.o
obj-y			+= cgroups.o
obj-y			+= log.o
obj-y			+= linux_kernel.o
obj-y			+= fs.o
obj-y			+= net.o
obj-y			+= security.o
obj-y			+= async.o
obj-y			+= util.o
obj-y			+= devnodes.o

cflags-y		+= -I.shipped/libnl/include/
cflags-y		+= -fPIC -Wa,--noexecstack -fno-stack-protector
cflags-so		+= -rdynamic -L.shipped/libnl/lib/.libs/ -lnl-route-3 -lnl-3

.SECONDARY:

ifneq ($(MAKECMDGOALS),clean)
incdeps := y
endif
