// Package headers lists well-known system headers used to classify
// #include directives.
package headers

// Builtin names a set of well-known headers.
type Builtin string

const (
	C     Builtin = "c"     // ISO C standard library, e.g. <stdio.h>.
	Cpp   Builtin = "cpp"   // ISO C++ standard library, e.g. <vector> or <cstdio>.
	Posix Builtin = "posix" // POSIX system headers, e.g. <unistd.h>.
)

// Contains reports whether name (the path between angle brackets) belongs to b.
// It reports false for unknown builtins.
func (b Builtin) Contains(name string) bool {
	set, ok := sets[b]
	if !ok {
		return false
	}
	_, ok = set[name]
	return ok
}

// Standard reports whether name (the path between angle brackets) is a C,
// C++ or POSIX standard header.
func Standard(name string) bool {
	return C.Contains(name) || Cpp.Contains(name) || Posix.Contains(name)
}

// Valid reports whether b is a known builtin set.
func (b Builtin) Valid() bool {
	_, ok := sets[b]
	return ok
}

var sets = map[Builtin]map[string]struct{}{
	C:     setOf(cHeaders),
	Cpp:   setOf(cppHeaders, cppCompatHeaders),
	Posix: setOf(posixHeaders),
}

func setOf(lists ...[]string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, list := range lists {
		for _, name := range list {
			set[name] = struct{}{}
		}
	}
	return set
}

// cHeaders is the C standard library up to C23.
var cHeaders = []string{
	"assert.h", "complex.h", "ctype.h", "errno.h", "fenv.h", "float.h",
	"inttypes.h", "iso646.h", "limits.h", "locale.h", "math.h", "setjmp.h",
	"signal.h", "stdalign.h", "stdarg.h", "stdatomic.h", "stdbit.h",
	"stdbool.h", "stdckdint.h", "stddef.h", "stdint.h", "stdio.h",
	"stdlib.h", "stdnoreturn.h", "string.h", "tgmath.h", "threads.h",
	"time.h", "uchar.h", "wchar.h", "wctype.h",
}

// cppHeaders is the C++ standard library up to C++26.
var cppHeaders = []string{
	"algorithm", "any", "array", "atomic", "barrier", "bit", "bitset",
	"charconv", "chrono", "codecvt", "compare", "complex", "concepts",
	"condition_variable", "contracts", "coroutine", "debugging", "deque",
	"exception", "execution", "expected", "filesystem", "flat_map",
	"flat_set", "format", "forward_list", "fstream", "functional", "future",
	"generator", "hazard_pointer", "hive", "initializer_list",
	"inplace_vector", "iomanip", "ios", "iosfwd", "iostream", "istream",
	"iterator", "latch", "limits", "linalg", "list", "locale", "map",
	"mdspan", "memory", "memory_resource", "mutex", "new", "numbers",
	"numeric", "optional", "ostream", "print", "queue", "random", "ranges",
	"ratio", "rcu", "regex", "scoped_allocator", "semaphore", "set",
	"shared_mutex", "simd", "source_location", "span", "spanstream",
	"sstream", "stack", "stacktrace", "stdexcept", "stdfloat", "stop_token",
	"streambuf", "string", "string_view", "strstream", "syncstream",
	"system_error", "text_encoding", "thread", "tuple", "type_traits",
	"typeindex", "typeinfo", "unordered_map", "unordered_set", "utility",
	"valarray", "variant", "vector", "version",
}

// cppCompatHeaders are the C++ wrappers of the C standard library.
var cppCompatHeaders = []string{
	"cassert", "ccomplex", "cctype", "cerrno", "cfenv", "cfloat",
	"cinttypes", "ciso646", "climits", "clocale", "cmath", "csetjmp",
	"csignal", "cstdalign", "cstdarg", "cstdbool", "cstddef", "cstdint",
	"cstdio", "cstdlib", "cstring", "ctgmath", "ctime", "cuchar", "cwchar",
	"cwctype",
}

// posixHeaders is the POSIX.1-2017 header set without the C standard library.
var posixHeaders = []string{
	"aio.h", "arpa/inet.h", "cpio.h", "dirent.h", "dlfcn.h", "fcntl.h",
	"fmtmsg.h", "fnmatch.h", "ftw.h", "glob.h", "grp.h", "iconv.h",
	"langinfo.h", "libgen.h", "monetary.h", "mqueue.h", "ndbm.h",
	"net/if.h", "netdb.h", "netinet/in.h", "netinet/tcp.h", "nl_types.h",
	"poll.h", "pthread.h", "pwd.h", "regex.h", "sched.h", "search.h",
	"semaphore.h", "spawn.h", "strings.h", "stropts.h", "sys/ipc.h",
	"sys/mman.h", "sys/msg.h", "sys/resource.h", "sys/select.h",
	"sys/sem.h", "sys/shm.h", "sys/socket.h", "sys/stat.h",
	"sys/statvfs.h", "sys/time.h", "sys/times.h", "sys/types.h",
	"sys/uio.h", "sys/un.h", "sys/utsname.h", "sys/wait.h", "syslog.h",
	"tar.h", "termios.h", "ulimit.h", "unistd.h", "utime.h", "utmpx.h",
	"wordexp.h",
}
