# cloudevent

[![GoDoc](https://godoc.org/github.com/fnproject/cloudevent?status.svg)](https://godoc.org/github.com/fnproject/cloudevent)

Go implementation of http://cloudevents.io/ specification

Example:

```go
import (
  "net/http"
  "fmt"

  "github.com/fnproject/cloudevent"
)

func handler(w http.ResponseWriter, r *http.Request) {
  var ce cloudevent.CloudEvent
  err := ce.FromRequest(r)
  _ = err // handle err

  fmt.Fprintln(w, ce.ContentType, ce.Data)
}
```

__UNSTABLE NOTICE:__ this repo is likely to undergo API changes at the moment,
any feedback on what those should look like would be much appreciated! You're
of course welcome to use this meanwhile, we are :)

TODO: tag a release version off for 0.1.1 for vendoring
