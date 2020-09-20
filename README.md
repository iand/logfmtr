# logfmtr

An implementation of the [logr minimal logging API](github.com/go-logr/logr) that writes in [logfmt](https://www.brandur.org/logfmt) style.

[![Build Status](https://travis-ci.org/iand/logfmtr.svg?branch=master)](https://travis-ci.org/iand/logfmtr)
[![Go Report Card](https://goreportcard.com/badge/github.com/iand/logfmtr)](https://goreportcard.com/report/github.com/iand/logfmtr)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/iand/logfmtr)

## Getting Started

This is a no frills logging package that follows the [logr minimal logging API](github.com/go-logr/logr) 
and by default writes logs in [logfmt](https://www.brandur.org/logfmt) format, a line of space delimited
key/value pairs.

```
level=0 logger=MyName msg=hello ts=2020-09-20T14:31:10.905267839Z user=you val1=1 val2=map[k:1]
level=1 logger=MyName msg="you should see this" ts=2020-09-20T14:31:10.905279546Z user=you
level=0 logger=MyName msg="uh oh" ts=2020-09-20T14:31:10.905288008Z error=<nil> user=you trouble=true reasons="[0.1 0.11 3.14]"
level=0 logger=MyName msg=goodbye ts=2020-09-20T14:31:10.905291479Z error="an error occurred" user=you code=-1
```

A more human friendly output format is also provided which can be configured using the `humanize` option:

```
0 info  | 14:31:10.905297 | hello                          logger=MyName user=you val1=1 val2=map[k:1]
1 info  | 14:31:10.905302 | you should see this            logger=MyName user=you
0 error | 14:31:10.905307 | uh oh                          logger=MyName error=<nil> user=you trouble=true reasons="[0.1 0.11 3.14]"
0 error | 14:31:10.905311 | goodbye                        logger=MyName error="an error occurred" user=you code=-1
```

Loggers defer applying their configuration until they are used. The logger is instantiated when
either Info, Error or Enabled is called. At that point the logger will read and use any options set
from a prior call to UseOptions. 


```Go
package main

import (
    "github.com/iand/logfmtr"
)

func main() {
    // Set options that all loggers will be based on
    opts := logfmtr.DefaultOptions()
    opts.Humanize = true
    opts.AddCaller = true
    logfmtr.UseOptions(opts)
}

```

Any new loggers will use the options set in `main` when they first start logging.

```Go
package worker

import (
    "github.com/iand/logfmtr"
)

// Create the logger
var logger = logfmtr.New().WithName("worker")


func doWork() }
    // Logger is instantiated with the options set earlier
    logger.Info("the sun is shining")
}
```

Loggers can be created with specific configuration by using `NewWithOptions`:

```Go
func otherWork() }
    opts := logfmtr.DefaultOptions()
    opts.Writer = os.Stderr

    logger := logfmtr.NewWithOptions(opts)
    logger.Info("important system information")
}
```

Several predefined keys are used when writing logs in logfmt style:

 * **msg** - the log message
 * **error** - error message passed to the `Error` method
 * **logger** - the name of the logger writing the log entry
 * **level** - the verbosity level of the logger writing the log entry
 * **caller** - filename and line number of the origin of the log entry

## Author

* [Ian Davis](http://github.com/iand) - <http://iandavis.com/>

## License

This is free and unencumbered software released into the public domain. Anyone is free to 
copy, modify, publish, use, compile, sell, or distribute this software, either in source 
code form or as a compiled binary, for any purpose, commercial or non-commercial, and by 
any means. For more information, see <http://unlicense.org/> or the 
accompanying [`UNLICENSE`](UNLICENSE) file.
