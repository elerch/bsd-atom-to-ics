package bsdatomtoics
 
import (
   "flag";
   "fmt";
   "net/http";
   "io";
   "os";
   "encoding/xml";
   "bytes";
   "strings";
   "time";
)

var debugFlag = flag.Bool("debug", false, "turns debugging messaging on");

const queryURI = "https://www.beaverton.k12.or.us/_vti_bin/BSD.Extranet/Syndication.svc/GetDistrictCalendarFeed?format=atom"
//const outputfmt = "%s \u27BE %s\n"
const outputfmt = "%s\n"
 
// Atom Feed Data Structure
 
type Feed struct {
   XMLName   xml.Name   "http://www.w3.org/2005/Atom feed";
   Title   string `xml:"title"`;
   Id   string `xml:"id"`;
   Link   []Link `xml:"link"`;
   Updated   Time `xml:"updated"`;
   Entry   []Entry `xml:"entry"`;
}
 
type Entry struct {
   Title   string `xml:"title"`;
   Id   string `xml:"id"`;
   Link   []Link `xml:"link"`;
   Updated   Time `xml:"updated"`;
   Content   string `xml:"content"`;
}
 
type Link struct {
   Rel   string   "attr";
   Href   string   "attr";
}

type Text struct {
   Type   string   "attr";
   Body   string   "chardata";
}
 
type Time string
 
func FetchBytes(debug bool) []byte {
   var rc []byte;
   r, err := http.Get(queryURI);
   if err == nil {
      fmt.Fprintf(os.Stderr, "\nRequest complete\n");
      if r.StatusCode == http.StatusOK {
         rc = readerToBytes(r.Body);
      } else {
         fmt.Fprintf(os.Stderr,
            "Error on request. Status: %s\n", r.Status)
      }
      r.Body.Close();
   } else {
      fmt.Fprintf(os.Stderr, "%v\n", err)
   }
   return rc;
}

func readerToBytes(r io.Reader) []byte {
   buf := new(bytes.Buffer);
   buf.ReadFrom(r);
   return buf.Bytes();
}
 
func AtomToICS(bytes []byte, writer io.Writer, debug bool) {
   var bsd Feed;
   err := xml.Unmarshal(bytes, &bsd);
 
   if (debug){
      fmt.Fprintf(os.Stderr, "Title: %s\n", bsd.Title);
      fmt.Fprintf(os.Stderr, "Id: %s\n", bsd.Id);
      fmt.Fprintf(os.Stderr, "Last updated: %v\n", len(bsd.Updated));
      fmt.Fprintf(os.Stderr, "Entry count after unmarshal: %v\n", len(bsd.Entry));
   }
   if err == nil {
      fmt.Printf("BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//BSDATOMTOICS v1.0//EN\n");
      for i := 0; i < len(bsd.Entry); i++ {
         fmt.Fprintf(writer, "BEGIN:VEVENT\n");
         fmt.Fprintf(writer, "SUMMARY:%s\n", bsd.Entry[i].Title);
         parseStartEndLocation(bsd.Entry[i].Content);
         start, end, location := parseStartEndLocation(bsd.Entry[i].Content);
         fmt.Fprintf(writer, "DTSTART:%s\n", start);
         fmt.Fprintf(writer, "DTEND:%s\n", end);
         fmt.Fprintf(writer, "LOCATION:%s\n", location);
         fmt.Fprintf(writer, "END:VEVENT\n");
      }
      fmt.Fprintf(writer, "END:VCALENDAR\n");
   } else {
      fmt.Fprintf(os.Stderr, "Unable to parse the Atom feed (%v)\n", err)
   }
}

func parseStartEndLocation(content string) (string, string, string) {
   //Event Time: 3/23/2015 12:00:00 PM - 3/27/2015 1:00:00 PM  Location:  Spring Break - Schools closed
   //fmt.Fprintf(os.Stderr, "Raw input: '%s'\n", content);
   strippedContent := strings.Replace(content, "\n", "", -1);
   reallyStrippedContent := strings.Replace(strippedContent, "\r", "", -1);
   eventTimeRemoved := strings.TrimLeft(strings.Replace(reallyStrippedContent, "Event Time: ", "", 1), " ");
   fmt.Fprintf(os.Stderr, "After removal: '%s'\n", eventTimeRemoved);
   timeFromLocation := strings.SplitAfterN(eventTimeRemoved, "Location: ", 2);
   time := timeFromLocation[0];
   location := timeFromLocation[1];
   startFromEnd := strings.SplitAfterN(time, " - ", 2);
   return toUTC(strings.Replace(startFromEnd[0], " - ", "", 1)),
          toUTC(strings.TrimRight(strings.Replace(startFromEnd[1], "Location: ", "", 1), " ")),
          strings.Trim(location, " ");
}

func toUTC(timeStr string) string {
   loc, _ := time.LoadLocation("America/Los_Angeles");
   //fmt.Fprintf(os.Stderr, "Input: |%s|\n", timeStr);
   //Mon Jan 2 15:04:05 -0700 MST 2006
   shortForm := "1/2/2006 3:04:05 PM";
   timeVal, _ := time.ParseInLocation(shortForm, timeStr, loc);
   return timeVal.UTC().Format("20060102T030400Z");
}
 
func dump(bytesToDump []byte) {
   buf := bytes.NewBuffer(bytesToDump);
   fmt.Printf(buf.String())
}
 
func main() {
   flag.Parse();
   //dump(fetchBytes(queryURI));
   // start, end, location := parseStartEndLocation("Event Time: 3/23/2015 12:00:00 PM - 3/27/2015 1:00:00 PM  Location:  Spring Break - Schools closed");
   // fmt.Fprintf(os.Stderr, "Start: %s\n", start);
   // fmt.Fprintf(os.Stderr, "End: %s\n", end);
   // fmt.Fprintf(os.Stderr, "Location: %s\n", location);
   AtomToICS(FetchBytes(*debugFlag), os.Stdout, *debugFlag);
}
