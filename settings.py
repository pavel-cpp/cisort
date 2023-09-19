from typing import Final

CPP_STD_LIBS: Final[dict[str, list[str]]] = {
    'io': [
        'fstream',
        'iomanip',
        'ios',
        'iosfwd',
        'iostream',
        'ostream',
        'istream',
        'sstream',
        'streambuf',
    ],
    'multithreading': [
        'atomic',
        'condition_variable',
        'future',
        'mutex',
        'shared_mutex',
        'thread',
    ],
    'other': [
        'algorithm',
        'bitset',
        'chrono',
        'codecvt',
        'complex',
        'exception',
        'functional',
        'initializer_list',
        'iterator',
        'limits',
        'locale',
        'memory',
        'new',
        'numeric',
        'random',
        'ratio',
        'regex',
        'stdexcept',
        'string',
        'system_error',
        'tuple',
        'typeindex',
        'typeinfo',
        'type_traits',
        'utility',
        'valarray',
    ],
    'STL': [
        'array',
        'deque',
        'forward_list',
        'list',
        'vector',
        'map',
        'set',
        'multimap',
        'multiset',
        'unordered_map',
        'unordered_set',
        'unordered_multimap',
        'unordered_multiset',
        'stack',
        'priority_queue',
        'queue',
    ],
    'clib': [
        'cassert',
        'cctype',
        'cerrno',
        'cfenv',
        'cfloat',
        'cinttypes',
        'ciso646',
        'climits',
        'clocale',
        'cmath',
        'csetjmp',
        'csignal',
        'cstdarg',
        'cstdbool',
        'cstddef',
        'cstdint',
        'cstdio',
        'cstdlib',
        'cstring',
        'ctgmath',
        'ctime',
        'cuchar',
        'cwchar',
        'cwctype',
    ]
}

C_STD_LIBS: Final[list[str]] = [
    'assert.h',
    'complex.h',
    'ctype.h',
    'errno.h',
    'fenv.h',
    'float.h',
    'inttypes.h',
    'iso646.h',
    'limits.h',
    'locale.h',
    'math.h',
    'setjmp.h',
    'signal.h',
    'stdalign.h',
    'stdarg.h',
    'stdatomic.h',
    'stdbool.h',
    'stddef.h',
    'stdint.h',
    'stdio.h',
    'stdlib.h',
    'stdnoreturn.h',
    'string.h',
    'tgmath.h',
    'threads.h',
    'time.h',
    'uchar.h',
    'wchar.h',
    'wctype.h',
]

FLAGS: Final = (
    '-h',
    '--help',
    '-ls',
    '-r'
)
