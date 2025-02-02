#include <stddef.h>

struct session {
    struct session *next;
    char *name;
    struct window *windows;
};

struct window {
    struct window *next;
    char *name;
    struct command *commands;
};

struct command {
    struct command *next;
    char *text;
};

void bail(const char *msg);
void *bail_alloc(size_t size);
char *bail_strdup(const char *s);

void add_session(struct session **sessions, char *name, struct window *windows);
void add_window(struct window **window, char *name, struct command *commands);
void add_command(struct command **command, char *text);

void destroy_sessions(struct session **sessions);
void destroy_windows(struct window **windows);
void destroy_commands(struct command **commands);

int search_directory(const char *dir_path, const char *target_name);
