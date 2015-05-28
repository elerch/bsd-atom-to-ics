package bsdatomtoics

import (
  "fmt"
  "net/http"
  "io"
  "io/ioutil"
  "os"
  "encoding/xml"
  "strings"
  "time"
)

const queryStart = "https://www.beaverton.k12.or.us/"
const queryEnd = "_vti_bin/BSD.Extranet/Syndication.svc/"
//District: "https://www.beaverton.k12.or.us/_vti_bin/BSD.Extranet/Syndication.svc/GetDistrictCalendarFeed?format=atom"
//JW "https://www.beaverton.k12.or.us/schools/jacob-wismer/_vti_bin/BSD.Extranet/Syndication.svc/GetSchoolEventsFeed?format=atom"
const outputfmt = "%s\n"

// Atom Feed Data Structure

type Feed struct {
  XMLName xml.Name "http://www.w3.org/2005/Atom feed"
  Title   string   `xml:"title"`
  Id      string   `xml:"id"`
  Link    []Link   `xml:"link"`
  Updated Time     `xml:"updated"`
  Entry   []Entry  `xml:"entry"`
}

type Entry struct {
  Title   string `xml:"title"`
  Id      string `xml:"id"`
  Link    []Link `xml:"link"`
  Updated Time   `xml:"updated"`
  Content string `xml:"content"`
}

type Link struct {
  Rel    string   "attr"
  Href   string   "attr"
}

type Text struct {
  Type   string   "attr"
  Body   string   "chardata"
}

type Time string

func FetchBytes() ([]byte, error) {
  return FetchBytesWith(http.DefaultClient, "")
}

func FetchBytesWith(client *http.Client, school string) ([]byte, error) {
   finalUri := queryStart;
   if school != "" { //jacob-wismer
      finalUri = finalUri + "schools/" + school + "/"
   }
   finalUri = finalUri + queryEnd
   if school == "" {
      finalUri = finalUri + "GetDistrictCalendarFeed?format=atom"
   } else {
      finalUri = finalUri + "GetSchoolEventsFeed?format=atom"
   } 
   
   r, err := client.Get(finalUri)
   if (err != nil || r.StatusCode != 200) {
      return nil, err
   }
   rc, err := ioutil.ReadAll(r.Body);
   defer r.Body.Close();
   if (err != nil) { return nil, err }
   return rc, nil;
}

func AtomToICS(bytes []byte, writer io.Writer, debug bool) {
  var bsd Feed

  if (bytes == nil || len(bytes) == 0) {
    fmt.Fprintf(writer, "no bytes received in input to AtomToICS")
    return
  }
  err := xml.Unmarshal(bytes, &bsd)

  if (debug){
    fmt.Fprintf(os.Stderr, "Title: %s\n", bsd.Title)
    fmt.Fprintf(os.Stderr, "Id: %s\n", bsd.Id)
    fmt.Fprintf(os.Stderr, "Last updated: %v\n", len(bsd.Updated))
    fmt.Fprintf(os.Stderr, "Entry count after unmarshal: %v\n", len(bsd.Entry))
  }
  if err == nil {
    fmt.Fprintf(writer, "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//BSDATOMTOICS v1.0//EN\r\n")
    for i := 0; i < len(bsd.Entry); i++ {
      fmt.Fprintf(writer, "BEGIN:VEVENT\r\n")
      fmt.Fprintf(writer, "SUMMARY:%s\r\n", bsd.Entry[i].Title)
      startTime, endTime, location := parseStartEndLocation(bsd.Entry[i].Content, debug)
      start, end := formatTimeForICS(startTime, endTime)
      fmt.Fprintf(writer, "DTSTART%s\r\n", start)
      fmt.Fprintf(writer, "DTEND%s\r\n", end)
      fmt.Fprintf(writer, "LOCATION:%s\r\n", location)
      fmt.Fprintf(writer, "END:VEVENT\r\n")
    }
    fmt.Fprintf(writer, "END:VCALENDAR\r\n")
  } else {
    fmt.Fprintf(os.Stderr, "Unable to parse the Atom feed (%v)\n", err)
  }
}

func parseStartEndLocation(content string, debug bool) (time.Time, time.Time, string) {
  //Event Time: 3/23/2015 12:00:00 PM - 3/27/2015 1:00:00 PM  Location:  Spring Break - Schools closed
  //fmt.Fprintf(os.Stderr, "Raw input: '%s'\n", content)
  strippedContent := strings.Replace(content, "\n", "", -1)
  reallyStrippedContent := strings.Replace(strippedContent, "\r", "", -1)
  eventTimeRemoved := strings.TrimLeft(strings.Replace(reallyStrippedContent, "Event Time: ", "", 1), " ")
  if (debug) {
    fmt.Fprintf(os.Stderr, "After removal: '%s'\n", eventTimeRemoved)
  }
  timeFromLocation := strings.SplitAfterN(eventTimeRemoved, "Location: ", 2)
  time := timeFromLocation[0]
  location := timeFromLocation[1]
  startFromEnd := strings.SplitAfterN(time, " - ", 2)
  return toTime(strings.Replace(startFromEnd[0], " - ", "", 1)),
         toTime(strings.TrimRight(strings.Replace(startFromEnd[1], "Location: ", "", 1), " ")),
         strings.Trim(location, " ")
}

func toTime(timeStr string) time.Time {
  loc, _ := time.LoadLocation("America/Los_Angeles")
  //fmt.Fprintf(os.Stderr, "Input: |%s|\n", timeStr)
  //Mon Jan 2 15:04:05 -0700 MST 2006
  shortForm := "1/2/2006 3:04:05 PM"
  timeVal, _ := time.ParseInLocation(shortForm, timeStr, loc)
  return timeVal//.UTC().Format("20060102T030400Z")
}

func formatTimeForICS(start time.Time, end time.Time) (string, string) {
  // start/end should be in local time. We need to detect midnight->11:59pm 
  // entries and switch them to all day
  startHour, startMin, startSec := start.Clock()
  endHour, endMin, endSec := end.Clock()
  if !(startHour == 0 && startMin == 0 && startSec == 0 &&
     endHour == 23 && endMin == 59 && endSec == 0) {
       return ":" + start.UTC().Format("20060102T150400Z"),
              ":" + end.UTC().Format("20060102T150400Z")
  }
  // we're on an all day event - grab the start date and build the
  // correct string values
  startDay := ";VALUE=DATE:" + start.Format("20060102")
  endDay   := ";VALUE=DATE:" + end.Add(time.Duration(1) * time.Minute).Format("20060102")
  
  return startDay, endDay
}