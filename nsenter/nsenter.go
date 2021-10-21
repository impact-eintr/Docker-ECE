package nsenter

/*
#define _GNU_SOURCE
#include <sched.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>

//这里的__attribute__((constructor))指的是， 一旦这个包被引用，那么这个函数就会被自动执行
//类似于构造函数，会在程序一启动的时候运行
__attribute__((constructor)) void enter_namespace(void) {
	char *docker_pid;
	docker_pid = getenv("docker_pid");
	if (docker_pid) {
		//fprintf(stdout, "got docker_pid=%s\n", docker_pid);
	} else {
		//fprintf(stdout, "missing docker_pid env skip nsenter");
		return;
	}
	char *docker_cmd;
	docker_cmd = getenv("docker_cmd");
	if (docker_cmd) {
		//fprintf(stdout, "got docker_cmd=%s\n", docker_cmd);
	} else {
		//fprintf(stdout, "missing docker_cmd env skip nsenter");
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", docker_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);

		if (setns(fd, 0) == -1) {
			//fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			//fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(docker_cmd);
	exit(0);
	return;
}
*/
import "C"
