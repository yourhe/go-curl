#include <stdlib.h>
#include <curl/curl.h>

static char *string_array_index(char **p, int i) {
  return p[i];
}