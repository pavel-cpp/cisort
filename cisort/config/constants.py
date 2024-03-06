from typing import Final

INCLUDE_REGEXP: Final[str] = r'^#include\s*([<\"][^>\"]*[>\"])'

EXTERNAL: Final[str] = '// External'

LOCAL: Final[str] = '// Local'

STARTUP_MESSAGE: Final[str] = 'Started cisort'

FILE_EXTENTIONS: tuple = ('.cpp', '.c', '.hpp', '.h')