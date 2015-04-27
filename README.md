# bsd-atom-to-ics

This golang package converts the Beaverton School District atom feed exposed at:

https://www.beaverton.k12.or.us/_vti_bin/BSD.Extranet/Syndication.svc/GetDistrictCalendarFeed?format=atom

To an ICS feed consumable by Google Calendar, Outlook, etc.

To use, "go get" the package:
```
go get github.com/elerch/bsd-atom-to-ics
```

Then you may use it in an application. Here is a sample web server that listens on port 8080 and delivers the ICS:

```go
package main

import (
    "net/http";
    "github.com/elerch/bsd-atom-to-ics";
)

func handler(w http.ResponseWriter, r *http.Request) {
    bsdatomtoics.AtomToICS(bsdatomtoics.FetchBytes(false), w, false);
}

func main() {
    http.HandleFunc("/", handler);
    http.ListenAndServe(":8080", nil);
}
```
