# moai-lang-c - Working Examples

_Last updated: 2025-11-21_

## Example 1: String Manipulation with Memory Management

```c
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

char* string_reverse(const char *str) {
    if (!str) return NULL;

    int len = strlen(str);
    char *reversed = (char *)malloc(len + 1);
    if (!reversed) return NULL;

    for (int i = 0; i < len; i++) {
        reversed[i] = str[len - 1 - i];
    }
    reversed[len] = '\0';

    return reversed;  // Caller must free
}

int main() {
    char *result = string_reverse("Hello, World!");
    if (result) {
        printf("Reversed: %s\n", result);
        free(result);
    }
    return 0;
}
```

## Example 2: Dynamic Array with Bounds Checking

```c
#include <stdlib.h>
#include <string.h>

typedef struct {
    int *data;
    int size;
    int capacity;
} DynamicArray;

DynamicArray* array_create(int initial_capacity) {
    DynamicArray *arr = malloc(sizeof(DynamicArray));
    arr->data = malloc(initial_capacity * sizeof(int));
    arr->size = 0;
    arr->capacity = initial_capacity;
    return arr;
}

int array_append(DynamicArray *arr, int value) {
    if (arr->size >= arr->capacity) {
        // Resize: double capacity
        int new_capacity = arr->capacity * 2;
        int *new_data = realloc(arr->data, new_capacity * sizeof(int));
        if (!new_data) return -1;  // Realloc failed

        arr->data = new_data;
        arr->capacity = new_capacity;
    }

    arr->data[arr->size++] = value;
    return 0;
}

int array_get(const DynamicArray *arr, int index) {
    if (index < 0 || index >= arr->size) {
        fprintf(stderr, "Array index out of bounds\n");
        return -1;
    }
    return arr->data[index];
}

void array_destroy(DynamicArray *arr) {
    if (arr) {
        free(arr->data);
        free(arr);
    }
}
```

## Example 3: File Processing with Error Handling

```c
#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>

int count_lines(const char *filename) {
    FILE *file = fopen(filename, "r");
    if (!file) {
        fprintf(stderr, "Cannot open %s: %s\n", filename, strerror(errno));
        return -1;
    }

    int count = 0;
    int c;

    while ((c = fgetc(file)) != EOF) {
        if (c == '\n') count++;
    }

    if (ferror(file)) {
        fprintf(stderr, "Error reading file: %s\n", strerror(errno));
        fclose(file);
        return -1;
    }

    fclose(file);
    return count;
}
```

## Example 4: Binary Search in Sorted Array

```c
#include <stdio.h>

int binary_search(const int *arr, int size, int target) {
    int left = 0, right = size - 1;

    while (left <= right) {
        int mid = left + (right - left) / 2;  // Avoid overflow

        if (arr[mid] == target) {
            return mid;
        } else if (arr[mid] < target) {
            left = mid + 1;
        } else {
            right = mid - 1;
        }
    }

    return -1;  // Not found
}
```

## Example 5: Pointer Arithmetic and Array Operations

```c
#include <stdio.h>
#include <stdlib.h>

void print_array(int *arr, int size) {
    // Pointer arithmetic: arr[i] is equivalent to *(arr + i)
    for (int *p = arr; p < arr + size; p++) {
        printf("%d ", *p);
    }
    printf("\n");
}

int* array_copy(const int *src, int size) {
    int *dest = malloc(size * sizeof(int));
    if (!dest) return NULL;

    // Copy using memcpy for efficiency
    memcpy(dest, src, size * sizeof(int));

    return dest;
}

void array_reverse_inplace(int *arr, int size) {
    int *start = arr;
    int *end = arr + size - 1;

    while (start < end) {
        // Swap
        int temp = *start;
        *start = *end;
        *end = temp;

        start++;
        end--;
    }
}
```

## Example 6: Integer Limits and Safe Arithmetic

```c
#include <stdio.h>
#include <limits.h>

// Safe addition: check for overflow
int safe_add(int a, int b, int *result) {
    if (b > 0 && a > INT_MAX - b) {
        return -1;  // Overflow
    }
    if (b < 0 && a < INT_MIN - b) {
        return -1;  // Underflow
    }

    *result = a + b;
    return 0;
}

int main() {
    int result;
    if (safe_add(1000000, 2000000, &result) == 0) {
        printf("Sum: %d\n", result);
    } else {
        printf("Arithmetic overflow detected\n");
    }
    return 0;
}
```

---

_For advanced patterns (Linked Lists, Memory Pools, Threading), see SKILL.md Advanced Patterns section_
