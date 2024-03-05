<img src="https://i.imgur.com/LsH6pj8.gif" alt="example" width="500"/>

# C/С++ Includes Sort

This is a tool that allows you to quickly and
conveniently sort all libraries included in the project in
alphabetical order.

## Examples
<img src="https://i.imgur.com/GLgtBDH.gif" alt="example" width="500"/>

### Text before:
```c++
#include <iostream>
#include <vector>
#include <stdio.h>
#include <algorithm>

#include "mylib.h"
#include "b_lib.h"
#include "a_first_lib.h"

#include <external/lib/main.hpp>
#include <external/lib/abuse.hpp>
#include <external/lib/func.hpp>

int main(){
	std::cout << "Hello, World!" << std::endl;
}
```
### Text after:
```c++
#include <algorithm>
#include <iostream>
#include <stdio.h>
#include <vector>

#include "a_first_lib.h"
#include "b_lib.h"
#include "mylib.h"

#include <external/lib/abuse.hpp>
#include <external/lib/func.hpp>
#include <external/lib/main.hpp>

int main(){
	std::cout << "Hello, World!" << std::endl;
}
```

## Installation:

```python
pip install cisort
```

## Usage:
```cisort [flags] [path]```

Flags:\
        `-r` - recursive searching C/C++ files\
        `-ls` - show info about sorted files\
        `-h` `--help` - to get help


## Authors
[Anton Zemtsov](https://github.com/antonata-c/)
[Pavel Remdenok](https://github.com/pavel-cpp/)


