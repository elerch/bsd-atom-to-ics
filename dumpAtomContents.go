package main

import ( 
  "bsdatomtoics"
  "fmt"
  "os"
  "bytes"
)

func dump(bytesToDump []byte) {
   buf := bytes.NewBuffer(bytesToDump)
   fmt.Printf(buf.String())
}
 
func main() {
   atom, err := bsdatomtoics.FetchBytes()
   if (err != nil) { 
      fmt.Fprintf(os.Stderr, "Error fetching atom data: &s", err.Error())
      return
   }
   dump(atom)
}
