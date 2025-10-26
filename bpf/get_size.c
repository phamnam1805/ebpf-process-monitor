#include <stdio.h>
#include <stdbool.h>
#include <stddef.h>
#include <netinet/in.h>
#include <linux/types.h>
#include "bootstrap.h"

int main() {
    printf("pid: size = %lu, offset = %lu\n", sizeof(((struct event *)0)->pid), offsetof(struct event, pid));
    printf("ppid: size = %lu, offset = %lu\n", sizeof(((struct event *)0)->ppid), offsetof(struct event, ppid));
    printf("exit_code: size = %lu, offset = %lu\n", sizeof(((struct event *)0)->exit_code), offsetof(struct event, exit_code));
    printf("duration_ns: size = %lu, offset = %lu\n", sizeof(((struct event *)0)->duration_ns), offsetof(struct event, duration_ns));
    printf("comm: size = %lu, offset = %lu\n", sizeof(((struct event *)0)->comm), offsetof(struct event, comm));
    printf("filename: size = %lu, offset = %lu\n", sizeof(((struct event *)0)->filename), offsetof(struct event, filename));
    printf("exit_event: size = %lu, offset = %lu\n", sizeof(((struct event *)0)->exit_event), offsetof(struct event, exit_event));
    return 0;
}