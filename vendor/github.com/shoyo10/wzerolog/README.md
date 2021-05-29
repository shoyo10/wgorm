# wzerolog

## Getting Started

### Simple Logging Example

```go
package main

import (
	"github.com/rs/zerolog/log"
	"github.com/shoyo10/wzerolog"
)

func main() {
	wzerolog.Init(wzerolog.Config{
		AppID:        "helloworld",
		Env:          "local",
	})
	log.Debug().Msg("hello world")
}

// Output: {"level":"debug","app_id":"helloworld","env":"local","log_time":1621324677094,"caller":"/workspace/main.go:13","message":"hello world"}
```

### Pretty Output

```go
package main

import (
	"github.com/rs/zerolog/log"
	"github.com/shoyo10/wzerolog"
)

func main() {
	wzerolog.Init(wzerolog.Config{
		PrettyOutput: true,
		AppID:        "helloworld",
		Env:          "local",
	})
	log.Debug().Msg("hello world")
}

// Output: 2021/05/18 16:02:24 DBG main.go:14 > [ hello world ] app_id:helloworld env:local
```

### Set Log Level

```go
package main

import (
	"github.com/rs/zerolog/log"
	"github.com/shoyo10/wzerolog"
)

func main() {
	wzerolog.Init(wzerolog.Config{
		LogLevel:     1,
		PrettyOutput: true,
		AppID:        "helloworld",
		Env:          "local",
	})
	log.Debug().Msg("not show")
	log.Info().Msg("hello world")
}

// Output: 2021/05/18 16:04:38 INF main.go:16 > [ hello world ] app_id:helloworld env:local
```