# Advanced C Programming Patterns

**Focus**: Production-ready patterns, meta-programming, data structures
**Standards**: C17 with advanced techniques
**Last Updated**: 2025-11-22

---

## Pattern 1: Opaque Pointer Pattern (Module Encapsulation)

Create private data structures through opaque pointers:

```c
// public_api.h
typedef struct _PrivateData PrivateData;  // Forward declaration

PrivateData *create_data(void);
void destroy_data(PrivateData *data);
void set_value(PrivateData *data, int value);
int get_value(const PrivateData *data);

// public_api.c
#include "public_api.h"

typedef struct {
    int value;
    char buffer[256];
} PrivateData;

PrivateData *create_data(void) {
    PrivateData *p = malloc(sizeof(PrivateData));
    if (!p) return NULL;
    p->value = 0;
    return p;
}

void destroy_data(PrivateData *data) {
    free(data);
}

void set_value(PrivateData *data, int value) {
    if (data) data->value = value;
}

int get_value(const PrivateData *data) {
    return data ? data->value : 0;
}
```

**Benefits**: Hide implementation details, enable ABI stability, support versioning

---

## Pattern 2: Custom Hash Table Implementation

Efficient key-value storage:

```c
#define HASH_TABLE_SIZE 1024

typedef struct {
    char *key;
    void *value;
} HashEntry;

typedef struct {
    HashEntry *entries[HASH_TABLE_SIZE];
    size_t count;
} HashMap;

// Simple hash function
static unsigned int hash(const char *key) {
    unsigned int h = 0;
    for (const char *p = key; *p; p++) {
        h = h * 31 + *p;
    }
    return h % HASH_TABLE_SIZE;
}

HashMap *hashmap_create(void) {
    HashMap *map = calloc(1, sizeof(HashMap));
    return map;
}

int hashmap_put(HashMap *map, const char *key, void *value) {
    if (!map || !key) return -1;

    unsigned int idx = hash(key);
    HashEntry *entry = malloc(sizeof(HashEntry));
    if (!entry) return -1;

    entry->key = malloc(strlen(key) + 1);
    if (!entry->key) {
        free(entry);
        return -1;
    }

    strcpy(entry->key, key);
    entry->value = value;

    map->entries[idx] = entry;
    map->count++;
    return 0;
}

void *hashmap_get(const HashMap *map, const char *key) {
    if (!map || !key) return NULL;

    unsigned int idx = hash(key);
    HashEntry *entry = map->entries[idx];

    return (entry && strcmp(entry->key, key) == 0) ? entry->value : NULL;
}

void hashmap_destroy(HashMap *map) {
    if (!map) return;

    for (int i = 0; i < HASH_TABLE_SIZE; i++) {
        if (map->entries[i]) {
            free(map->entries[i]->key);
            free(map->entries[i]);
        }
    }
    free(map);
}
```

---

## Pattern 3: Variadic Macro for Type-Safe Callbacks

Enhanced flexibility and type safety:

```c
#define DEFINE_CALLBACK(NAME, RETURN_TYPE, ...) \
    typedef RETURN_TYPE (*NAME)(##__VA_ARGS__)

DEFINE_CALLBACK(IntCallback, int, int);
DEFINE_CALLBACK(VoidCallback, void, const char *);
DEFINE_CALLBACK(BoolCallback, int, void *, int);

// Usage
IntCallback my_callback = some_function;
VoidCallback log_callback = logger_function;
```

---

## Pattern 4: Arena Allocation with Alignment

Performance-optimized memory allocation:

```c
#include <stdint.h>

#define ALIGN_UP(size, align) (((size) + (align) - 1) & ~((align) - 1))

typedef struct {
    uint8_t *buffer;
    size_t capacity;
    size_t offset;
} Arena;

Arena *arena_create(size_t capacity) {
    Arena *arena = malloc(sizeof(Arena));
    if (!arena) return NULL;

    arena->buffer = malloc(capacity);
    if (!arena->buffer) {
        free(arena);
        return NULL;
    }

    arena->capacity = capacity;
    arena->offset = 0;
    return arena;
}

void *arena_alloc(Arena *arena, size_t size, size_t align) {
    if (!arena || size == 0) return NULL;

    size_t aligned_offset = ALIGN_UP(arena->offset, align);
    if (aligned_offset + size > arena->capacity) {
        return NULL;
    }

    void *ptr = &arena->buffer[aligned_offset];
    arena->offset = aligned_offset + size;
    return ptr;
}

void arena_reset(Arena *arena) {
    if (arena) arena->offset = 0;
}

void arena_destroy(Arena *arena) {
    if (!arena) return;
    free(arena->buffer);
    free(arena);
}
```

---

## Pattern 5: Polymorphism Through Function Pointers

OOP-like behavior in C:

```c
typedef struct {
    void (*init)(void *);
    void (*process)(void *, const void *);
    void (*cleanup)(void *);
    void (*destroy)(void *);
} VTable;

typedef struct {
    VTable *vtable;
    void *data;
} Object;

// Concrete implementation
typedef struct {
    int count;
    char buffer[256];
} FileProcessor;

void file_init(void *self) {
    FileProcessor *fp = (FileProcessor *)self;
    fp->count = 0;
}

void file_process(void *self, const void *input) {
    FileProcessor *fp = (FileProcessor *)self;
    fp->count++;
}

void file_cleanup(void *self) {
    FileProcessor *fp = (FileProcessor *)self;
    // cleanup resources
}

void file_destroy(void *self) {
    free(self);
}

VTable file_vtable = {
    .init = file_init,
    .process = file_process,
    .cleanup = file_cleanup,
    .destroy = file_destroy
};

// Usage
Object *create_file_processor(void) {
    Object *obj = malloc(sizeof(Object));
    obj->vtable = &file_vtable;
    obj->data = malloc(sizeof(FileProcessor));
    obj->vtable->init(obj->data);
    return obj;
}

void process_object(Object *obj, const void *input) {
    obj->vtable->process(obj->data, input);
}
```

---

## Pattern 6: Inline Assembly for Performance

SIMD and low-level optimizations:

```c
// SIMD sum with SSE2
#include <emmintrin.h>

int simd_sum(const int *arr, int n) {
    __m128i sum = _mm_setzero_si128();

    for (int i = 0; i < n; i += 4) {
        __m128i v = _mm_loadu_si128((__m128i *)&arr[i]);
        sum = _mm_add_epi32(sum, v);
    }

    // Horizontal sum
    __m128i s = sum;
    s = _mm_hadd_epi32(s, s);
    s = _mm_hadd_epi32(s, s);
    return _mm_cvtsi128_si32(s);
}

// Inline assembly example (x86-64)
int inline_asm_add(int a, int b) {
    int result;
    asm("addl %2, %0"
        : "=r" (result)
        : "0" (a), "r" (b));
    return result;
}
```

---

## Pattern 7: Macro Metaprogramming

Template-like functionality in C:

```c
// Container-of pattern
#define container_of(ptr, type, member) \
    ((type *)((char *)(ptr) - offsetof(type, member)))

typedef struct {
    int key;
    int value;
} KeyValue;

typedef struct {
    int data;
    KeyValue kv;
} Container;

Container *get_container(KeyValue *kv) {
    return container_of(kv, Container, kv);
}

// Generic list macro
#define LIST_FOR_EACH(node, head) \
    for ((node) = (head); (node); (node) = (node)->next)

typedef struct ListNode {
    void *data;
    struct ListNode *next;
} ListNode;

ListNode *head = NULL;
ListNode *node;
LIST_FOR_EACH(node, head) {
    printf("%p\n", node->data);
}
```

---

## Pattern 8: Thread-Safe Reference Counting

Safe memory management in concurrent code:

```c
#include <stdatomic.h>
#include <threads.h>

typedef struct {
    _Atomic(int) refcount;
    void *data;
    void (*destroy)(void *);
} RefCounted;

RefCounted *refcounted_create(void *data, void (*destroy)(void *)) {
    RefCounted *rc = malloc(sizeof(RefCounted));
    if (!rc) return NULL;

    atomic_init(&rc->refcount, 1);
    rc->data = data;
    rc->destroy = destroy;
    return rc;
}

RefCounted *refcounted_acquire(RefCounted *rc) {
    if (rc) atomic_fetch_add(&rc->refcount, 1);
    return rc;
}

void refcounted_release(RefCounted *rc) {
    if (!rc) return;

    if (atomic_fetch_sub(&rc->refcount, 1) == 1) {
        if (rc->destroy) rc->destroy(rc->data);
        free(rc);
    }
}

int refcounted_count(const RefCounted *rc) {
    return rc ? atomic_load(&rc->refcount) : 0;
}
```

---

## Pattern 9: Compile-Time Assertions

Catch errors at compile time:

```c
// Pre-C11 version
#define STATIC_ASSERT(expr) \
    extern int static_assertion_failed[(expr) ? 1 : -1]

// C11+ version
#define static_assert _Static_assert

// Usage
typedef struct {
    int x;
    int y;
} Point;

// Ensure structure has expected size
static_assert(sizeof(Point) == 8, "Point should be 8 bytes");

// Ensure enum values
enum Status {
    STATUS_OK = 0,
    STATUS_ERROR = 1
};
static_assert(STATUS_OK == 0, "STATUS_OK must be 0");
```

---

## Pattern 10: Error Handling Context

Structured error reporting:

```c
typedef struct {
    int code;
    char message[256];
    const char *file;
    int line;
} ErrorContext;

#define ERROR_CONTEXT_INIT() {.code = 0, .message = {0}, .file = NULL, .line = 0}

#define SET_ERROR(ctx, code, msg) do { \
    (ctx)->code = code; \
    strncpy((ctx)->message, msg, sizeof((ctx)->message) - 1); \
    (ctx)->file = __FILE__; \
    (ctx)->line = __LINE__; \
} while(0)

int process_data(const char *input, ErrorContext *error) {
    if (!input) {
        SET_ERROR(error, -1, "Input is NULL");
        return -1;
    }

    // Process
    return 0;
}

// Usage
ErrorContext err = ERROR_CONTEXT_INIT();
if (process_data(NULL, &err) < 0) {
    printf("Error at %s:%d: %s\n", err.file, err.line, err.message);
}
```

---

**Last Updated**: 2025-11-22 | **Standards**: C17 | **Production Ready**
