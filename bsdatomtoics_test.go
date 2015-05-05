package bsdatomtoics

import (
  "testing"
  "io/ioutil"
  "bytes"
  "fmt"
)

func TestNothing(t *testing.T) {
  atom, err := ioutil.ReadFile("capturedAtom.txt")
  if err != nil {
    t.Error(err)
    return
  }
  expected, err := ioutil.ReadFile("expectedICS.txt")
  if err != nil {
    t.Error(err)
    return
  }
  buffer := new(bytes.Buffer)
  AtomToICS(atom, buffer, false)
  if len(expected) != buffer.Len() {
    t.Error(fmt.Sprintf("Actual output is length %v, but expected %v", buffer.Len(), len(expected)))
    ioutil.WriteFile("actualICS.txt", buffer.Bytes(), 0644)
    return
  }
}
