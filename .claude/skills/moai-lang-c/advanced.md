# moai-lang-c - Advanced Patterns

_Last updated: 2025-11-21_

Advanced C programming patterns for professional systems development, concurrent programming, and low-level system integration.

## Advanced Patterns

### 1. **Linked List Implementation** (Memory & Pointer Mastery)

```c
typedef struct Node {
    int data;
    struct Node *next;
} Node;

Node* create_node(int data) {
    Node *node = (Node *)malloc(sizeof(Node));
    if (!node) return NULL;

    node->data = data;
    node->next = NULL;
    return node;
}

void append(Node **head, int data) {
    Node *new_node = create_node(data);
    if (!new_node) return;

    if (*head == NULL) {
        *head = new_node;
        return;
    }

    Node *current = *head;
    while (current->next != NULL) {
        current = current->next;
    }
    current->next = new_node;
}

void free_list(Node *head) {
    while (head) {
        Node *temp = head;
        head = head->next;
        free(temp);
    }
}
```

### 2. **Memory Pool Pattern** (Efficient Allocation)

```c
typedef struct MemoryPool {
    void **free_blocks;
    int block_size;
    int pool_size;
    int available;
} MemoryPool;

MemoryPool* create_pool(int block_size, int pool_size) {
    MemoryPool *pool = malloc(sizeof(MemoryPool));
    pool->block_size = block_size;
    pool->pool_size = pool_size;
    pool->available = pool_size;
    pool->free_blocks = malloc(pool_size * sizeof(void *));

    for (int i = 0; i < pool_size; i++) {
        pool->free_blocks[i] = malloc(block_size);
    }
    return pool;
}

void* pool_alloc(MemoryPool *pool) {
    if (pool->available <= 0) return NULL;
    return pool->free_blocks[--pool->available];
}

void pool_free(MemoryPool *pool, void *ptr) {
    if (pool->available < pool->pool_size) {
        pool->free_blocks[pool->available++] = ptr;
    }
}
```

### 3. **File I/O with Error Handling**

```c
#include <stdio.h>
#include <errno.h>
#include <string.h>

typedef struct {
    FILE *file;
    char *filename;
    int error;
} FileHandle;

FileHandle* file_open(const char *filename, const char *mode) {
    FileHandle *fh = malloc(sizeof(FileHandle));
    if (!fh) return NULL;

    fh->file = fopen(filename, mode);
    if (!fh->file) {
        fh->error = errno;
        fprintf(stderr, "Error opening %s: %s\n", filename, strerror(errno));
        return fh;
    }

    fh->filename = malloc(strlen(filename) + 1);
    strcpy(fh->filename, filename);
    fh->error = 0;
    return fh;
}

int file_read_lines(FileHandle *fh, char **lines, int max_lines) {
    int count = 0;
    char buffer[1024];

    while (count < max_lines && fgets(buffer, sizeof(buffer), fh->file)) {
        lines[count] = malloc(strlen(buffer) + 1);
        strcpy(lines[count], buffer);
        count++;
    }

    if (ferror(fh->file)) {
        fh->error = errno;
        return -1;
    }
    return count;
}

void file_close(FileHandle *fh) {
    if (fh) {
        if (fh->file) fclose(fh->file);
        if (fh->filename) free(fh->filename);
        free(fh);
    }
}
```

### 4. **Process Management** (fork/exec patterns)

```c
#include <unistd.h>
#include <sys/wait.h>
#include <stdlib.h>

int execute_command(const char *command, const char *args[]) {
    pid_t pid = fork();

    if (pid == -1) {
        perror("fork");
        return -1;
    }

    if (pid == 0) {
        execvp(command, (char * const *)args);
        perror("execvp");
        exit(EXIT_FAILURE);
    } else {
        int status;
        waitpid(pid, &status, 0);
        return WIFEXITED(status) ? WEXITSTATUS(status) : -1;
    }
}
```

### 5. **Mutex & Synchronization** (Thread-safe patterns)

```c
#include <pthread.h>

typedef struct {
    int value;
    pthread_mutex_t lock;
} ThreadSafeCounter;

ThreadSafeCounter* counter_create() {
    ThreadSafeCounter *c = malloc(sizeof(ThreadSafeCounter));
    c->value = 0;
    pthread_mutex_init(&c->lock, NULL);
    return c;
}

void counter_increment(ThreadSafeCounter *c) {
    pthread_mutex_lock(&c->lock);
    c->value++;
    pthread_mutex_unlock(&c->lock);
}

int counter_get(ThreadSafeCounter *c) {
    pthread_mutex_lock(&c->lock);
    int value = c->value;
    pthread_mutex_unlock(&c->lock);
    return value;
}

void counter_destroy(ThreadSafeCounter *c) {
    pthread_mutex_destroy(&c->lock);
    free(c);
}
```

### 6. **Macro & Preprocessor Patterns** (Meta-programming)

```c
// Reusable vector macro
#define DEFINE_VECTOR(Type) \
    typedef struct { \
        Type *data; \
        int size; \
        int capacity; \
    } Type##_Vector; \
    \
    Type##_Vector* Type##_vector_new(int capacity) { \
        Type##_Vector *v = malloc(sizeof(Type##_Vector)); \
        v->data = malloc(capacity * sizeof(Type)); \
        v->size = 0; \
        v->capacity = capacity; \
        return v; \
    } \
    \
    void Type##_vector_push(Type##_Vector *v, Type value) { \
        if (v->size >= v->capacity) return; \
        v->data[v->size++] = value; \
    }

DEFINE_VECTOR(int)
DEFINE_VECTOR(double)
```

---

_For basic C programming, see SKILL.md and examples.md_
_For development tools, see reference.md_
