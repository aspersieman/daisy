#include <yaml.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dirent.h>
#include <sys/stat.h>
#include <unistd.h>
#include "daisy.h"

// int main(int argc, char *argv[]) {
//     if (argc != 3) {
//         fprintf(stderr, "Usage: %s <directory> <filename>\n", argv[0]);
//         return EXIT_FAILURE;
//     }
//
//     const char *dir_path = argv[1];
//     const char *target_name = argv[2];
//
//     if (search_directory(dir_path, target_name)) {
//         printf("File %s found.\n", target_name);
//         return EXIT_SUCCESS;
//     } else {
//         printf("File %s not found.\n", target_name);
//         return EXIT_FAILURE;
//     }
// }

/* Helper to bail on error. */
void bail(const char *msg)
{
    fprintf(stderr, "%s\n", msg);
    exit(1);
}

/* Helper to allocate memory or bail. */
void *bail_alloc(size_t size)
{
    void *p = calloc(1, size);
    if (!p) {
        bail("out of memory");
    }
    return p;
}

/* Helper to copy a string or bail. */
char *bail_strdup(const char *s)
{
    char *c = strdup(s ? s : "");
    if (!c) {
        bail("out of memory");
    }
    return c;
}

void add_session(struct session **sessions, char *name, struct window *windows)
{
    /* Create session object. */
    struct session *f = bail_alloc(sizeof(*f));
    f->name = bail_strdup(name);
    f->windows = windows;

    /* Append to list. */
    if (!*sessions) {
        *sessions = f;
    } else {
        struct session *tail = *sessions;
        while (tail->next) {
            tail = tail->next;
        }
        tail->next = f;
    }
}

void add_window(struct window **windows, char *name, struct command *commands)
{
    /* Create window object. */
    struct window *w = bail_alloc(sizeof(*w));
    w->name = bail_strdup(name);

    /* Append to list. */
    if (!*windows) {
        *windows = w;
    } else {
        struct window *tail = *windows;
        while (tail->next) {
            tail = tail->next;
        }
        tail->next = w;
    }
}

void add_command(struct command **commands, char *text)
{
    /* Create command object. */
    struct command *c = bail_alloc(sizeof(*c));

    /* Append to list. */
    if (!*commands) {
        *commands = c;
    } else {
        struct command *tail = *commands;
        while (tail->next) {
            tail = tail->next;
        }
        tail->next = c;
    }
}

int search_directory(const char *dir_path, const char *target_name) {
    DIR *dir = opendir(dir_path);
    struct dirent *entry;
    struct stat statbuf;
    char path[1024];

    if (dir == NULL) {
        perror("opendir");
        return 0;
    }

    while ((entry = readdir(dir)) != NULL) {
        // Skip . and .. directories
        if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) {
            continue;
        }

        // Build the full path of the entry
        snprintf(path, sizeof(path), "%s/%s", dir_path, entry->d_name);

        // Get the file's information
        if (stat(path, &statbuf) == -1) {
            perror("stat");
            closedir(dir);
            return 0;
        }

        // If it's a directory, recurse into it
        if (S_ISDIR(statbuf.st_mode)) {
            if (search_directory(path, target_name)) {
                closedir(dir);
                return 1;
            }
        } else if (S_ISREG(statbuf.st_mode)) {
            // If it's a file, check if the name matches
            if (strcmp(entry->d_name, target_name) == 0) {
                printf("Found: %s\n", path);
                closedir(dir);
                return 1;
            }
        }
    }

    closedir(dir);
    return 0;
}

/* Set environment variable DEBUG=1 to enable debug output. */
int debug = 0;

/* yaml_* functions return 1 on success and 0 on failure. */
enum status {
    SUCCESS = 1,
    FAILURE = 0
};

/* Our example parser states. */
enum state {
    STATE_START,    /* start state */
    STATE_STREAM,   /* start/end stream */
    STATE_DOCUMENT, /* start/end document */
    STATE_SECTION,  /* top level */

    STATE_SLIST,    /* session list */
    STATE_SVALUES,  /* session key-value pairs */
    STATE_SKEY,     /* session key */
    STATE_SNAME,    /* session name value */

    STATE_WLIST,    /* window list */
    STATE_WVALUES,  /* window key-value pairs */
    STATE_WKEY,     /* window key */
    STATE_WNAME,    /* window name */

    STATE_CLIST,    /* command list */
    STATE_CVALUES,  /* command key-value pairs */
    STATE_CKEY,     /* command key */

    STATE_STOP      /* end state */
};

/* Our application parser state data. */
struct parser_state {
    enum state state;        /* The current parse state */
    struct session s;        /* Session data elements. */
    struct window  w;        /* Window data elements. */
    struct command c;        /* Command data elements. */
    struct session *slist;   /* List of 'session' objects. */
    struct window *wlist;    /* List of 'window' objects. */
    struct command *clist;   /* List of 'command' objects. */
};

/*
 * Consume yaml events generated by the libyaml parser to
 * import our data into raw c data structures. Error processing
 * is keep to a mimimum since this is just an example.
 */
int consume_event(struct parser_state *s, yaml_event_t *event)
{
    char *value;

    if (debug) {
        printf("state=%d event=%d\n", s->state, event->type);
    }
    switch (s->state) {
    case STATE_START:
        switch (event->type) {
        case YAML_STREAM_START_EVENT:
            s->state = STATE_STREAM;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

     case STATE_STREAM:
        switch (event->type) {
        case YAML_DOCUMENT_START_EVENT:
            s->state = STATE_DOCUMENT;
            break;
        case YAML_STREAM_END_EVENT:
            s->state = STATE_STOP;  /* All done. */
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

     case STATE_DOCUMENT:
        switch (event->type) {
        case YAML_MAPPING_START_EVENT:
            s->state = STATE_SECTION;
            break;
        case YAML_DOCUMENT_END_EVENT:
            s->state = STATE_STREAM;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_SECTION:
        switch (event->type) {
        case YAML_SCALAR_EVENT:
            value = (char *)event->data.scalar.value;
            if (strcmp(value, "session") == 0) {
               s->state = STATE_SLIST;
            } else {
               fprintf(stderr, "Unexpected scalar: %s\n", value);
               return FAILURE;
            }
            break;
        case YAML_DOCUMENT_END_EVENT:
            s->state = STATE_STREAM;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_SLIST:
        switch (event->type) {
        case YAML_SEQUENCE_START_EVENT:
            s->state = STATE_SVALUES;
            break;
        case YAML_MAPPING_END_EVENT:
            s->state = STATE_SECTION;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_SVALUES:
        switch (event->type) {
        case YAML_MAPPING_START_EVENT:
            s->state = STATE_SKEY;
            break;
        case YAML_SEQUENCE_END_EVENT:
            s->state = STATE_SLIST;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_SKEY:
        switch (event->type) {
        case YAML_SCALAR_EVENT:
            value = (char *)event->data.scalar.value;
            if (strcmp(value, "name") == 0) {
                s->state = STATE_SNAME;
            } else if (strcmp(value, "windows") == 0) {
                s->state = STATE_WLIST;
            } else {
                fprintf(stderr, "Unexpected key: %s\n", value);
                return FAILURE;
            }
            break;
        case YAML_MAPPING_END_EVENT:
            add_session(&s->slist, s->s.name, s->wlist);
            free(s->s.name);
            memset(&s->s, 0, sizeof(s->s));
            s->wlist = NULL;
            s->state = STATE_SVALUES;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_SNAME:
        switch (event->type) {
        case YAML_SCALAR_EVENT:
            if (s->s.name) {
                fprintf(stderr, "Warning: duplicate 'name' key.\n");
                free(s->s.name);
            }
            s->s.name = bail_strdup((char *)event->data.scalar.value);
            s->state = STATE_SKEY;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;


    case STATE_WLIST:
        switch (event->type) {
        case YAML_SEQUENCE_START_EVENT:
            s->state = STATE_WVALUES;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_WVALUES:
        switch (event->type) {
        case YAML_MAPPING_START_EVENT:
            s->state = STATE_WKEY;
            break;
        case YAML_SEQUENCE_END_EVENT:
            s->state = STATE_SKEY;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_WKEY:
        switch (event->type) {
        case YAML_SCALAR_EVENT:
            value = (char *)event->data.scalar.value;
            if (strcmp(value, "name") == 0) {
                s->state = STATE_WNAME;
            } else if (strcmp(value, "commands") == 0) {
                s->state = STATE_CLIST;
            } else {
                fprintf(stderr, "Unexpected key: %s\n", value);
                return FAILURE;
            }
            break;
        case YAML_MAPPING_END_EVENT:
            add_window(&s->wlist, s->w.name, s->clist);
            free(s->w.name);
            memset(&s->w, 0, sizeof(s->w));
            s->state = STATE_WVALUES;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_WNAME:
        switch (event->type) {
        case YAML_SCALAR_EVENT:
            if (s->w.name) {
                fprintf(stderr, "Warning: duplicate 'name' key.\n");
                free(s->w.name);
            }
            s->w.name = bail_strdup((char *)event->data.scalar.value);
            s->state = STATE_WKEY;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_CLIST:
        switch (event->type) {
        case YAML_SEQUENCE_START_EVENT:
            s->state = STATE_CVALUES;
            break;
        case YAML_MAPPING_END_EVENT:
            s->state = STATE_SECTION;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_CVALUES:
        switch (event->type) {
        case YAML_MAPPING_START_EVENT:
            s->state = STATE_CKEY;
            break;
        case YAML_SEQUENCE_END_EVENT:
            s->state = STATE_CLIST;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_CKEY:
        switch (event->type) {
        case YAML_SCALAR_EVENT:
            value = (char *)event->data.scalar.value;
            break;
        case YAML_MAPPING_END_EVENT:
            add_command(&s->clist, s->s.name);
            memset(&s->c, 0, sizeof(s->c));
            s->clist = NULL;
            s->state = STATE_CVALUES;
            break;
        default:
            fprintf(stderr, "Unexpected event %d in state %d.\n", event->type, s->state);
            return FAILURE;
        }
        break;

    case STATE_STOP:
        break;
    }
    return SUCCESS;
}

void destroy_sessions(struct session **sessions)
{
    for (struct session *s = *sessions; s; s = *sessions) {
        *sessions = s->next;
        free(s->name);
        destroy_windows(&s->windows);
        free(s);
    }
}

void destroy_windows(struct window **windows)
{
    for (struct window *w = *windows; w; w = *windows) {
        *windows = w->next;
        free(w->name);
        destroy_commands(&w->commands);
        free(w);
    }
}

void destroy_commands(struct command **commands)
{
    for (struct command *c = *commands; c; c = *commands) {
        *commands = c->next;
        free(c);
    }
}

int main(int argc, char *argv[])
{
    int code;
    enum status status;
    struct parser_state state;
    yaml_parser_t parser;

    if (getenv("DEBUG")) {
        debug = 1;
    }

    memset(&state, 0, sizeof(state));
    state.state = STATE_START;
    yaml_parser_initialize(&parser);
    yaml_parser_set_input_file(&parser, stdin);
    do {
        yaml_event_t event;

        status = yaml_parser_parse(&parser, &event);
        if (status == FAILURE) {
            fprintf(stderr, "yaml_parser_parse error\n");
            code = EXIT_FAILURE;
            goto done;
        }
        status = consume_event(&state, &event);
        yaml_event_delete(&event);
        if (status == FAILURE) {
            fprintf(stderr, "consume_event error\n");
            code = EXIT_FAILURE;
            goto done;
        }
    } while (state.state != STATE_STOP);

    /* Output the parsed data. */
    for (struct session *s = state.slist; s; s = s->next) {
        printf("session: name=%s\n", s->name);
        for (struct window *w = s->windows; w; w = w->next) {
            printf("  window: name=%s\n", w->name);
        }
    }
    code = EXIT_SUCCESS;

done:
    free(state.s.name);
    free(state.w.name);
    destroy_sessions(&state.slist);
    destroy_windows(&state.wlist);
    yaml_parser_delete(&parser);
    return code;
}
