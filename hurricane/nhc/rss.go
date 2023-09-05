package nhc

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mmcdole/gofeed"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const mphToKts = 0.868976

var publicAdvRegex = regexp.MustCompile(`^(?:(?:Post-)Tropical (?:Storm|Depression|Cyclone|Hurricane)) (.*) Public Advisory Number (.*)`)
var publicAdvRegexOfDoom = regexp.MustCompile(`(?s).*(AL[0-9][0-9][0-9][0-9][0-9][0-9])\n(.*)\n.\n.*\n.\n.*SUMMARY.*LOCATION\.\.\.([0-9]?[0-9]?[0-9]?\.[0-9]?[0-9]?)([NS]) ([0-9]?[0-9]?[0-9]?\.[0-9]?[0-9]?)([EW]).*MAXIMUM.*\.\.\.([0-9]?[0-9]?[0-9]?) MPH.*PRESENT.*OR ([0-9]?[0-9]?[0-9]?).*AT ([0-9]?[0-9]?[0-9]?) MPH.*MINIMUM CENTRAL PRESSURE\.\.\.([0-9]?[0-9]?[0-9]?[0-9]?) MB.*DISCUSSION AND OUTLOOK\n(?:[-]*\n)(.*)(?:\n[[:blank:]]\n[[:blank:]]\n).*(?:\n[[:blank:]]\n[[:blank:]]\n).*`)
var graphicsRegex = regexp.MustCompile(`(?:Tropical (?:Storm|Depression)|Hurricane) (.*) Graphics`)

// PubSubMessage is the payload of a Pub/Sub event. Please refer to the docs for
// additional information regarding Pub/Sub events.
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// PubSubEntry consumes a Pub/Sub message.
func PubSubEntry(ctx context.Context, m PubSubMessage) error {
	log.Println(string(m.Data))
	ParseFeed()
	return nil
}

type StormFeedInfo struct {
	Name            string    `json:"name,omitempty"`
	StormID         string    `json:"storm_id,omitempty"`
	AdvNumber       int       `json:"adv_number,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
	LatY            float64   `json:"lat_y,omitempty"`
	LonX            float64   `json:"lon_x,omitempty"`
	BearingDeg      float64   `json:"bearing_deg,omitempty"`
	ForwardSpeedKts float64   `json:"forward_speed_kts,omitempty"`
	VMaxKts         float64   `json:"v_max_kts,omitempty"`
	MinCpMb         float64   `json:"min_cp_mb,omitempty"`
	Discussion      string    `json:"discussion,omitempty"`
	Graphics        []string  `json:"graphics,omitempty"`
	Sources         []string  `json:"sources,omitempty"`
	Raw             string    `json:"raw,omitempty"`
}

func (i *StormFeedInfo) SetGraphics(g []string) {
	i.Graphics = g
}

func (i *StormFeedInfo) SetSources(g []string) {
	i.Sources = g
}

func isPublicAdv(adv *gofeed.Item) bool {
	return publicAdvRegex.MatchString(adv.Title)
}

func parsePublicAdv(adv *gofeed.Item) (info StormFeedInfo) {
	name := publicAdvRegex.ReplaceAllString(adv.Title, "$1")
	advNumber, _ := strconv.Atoi(publicAdvRegex.ReplaceAllString(adv.Title, "$2"))
	allMatches := publicAdvRegexOfDoom.FindAllStringSubmatch(adv.Description, -1)
	matchVars := allMatches[0][1:]

	stormID := matchVars[0]

	// Second Note: It seems std lib Go can _only_ handle UTC and local timezone, so switching to tracking
	// by published date with min/sec zero'd for ts tracking
	// https://stackoverflow.com/questions/25368415/how-to-properly-parse-timezone-codes
	// Hacky parsing because golang can't deal with the 500/1100 PM time format
	// https://github.com/golang/go/issues/12919 (reporter even points to the NWS too lol)
	// Example failing TS: "500 AM AST Mon Sep 20 2021"
	//rawTsString := matchVars[1]
	//splits := strings.SplitN(rawTsString, " ", 2)
	//timePart := splits[0]
	//var hourPart string
	//var minPart string
	//if len(timePart) == 3 {
	//	hourPart = timePart[0:1]
	//	minPart = timePart[1:3]
	//} else if len(timePart) == 4 {
	//	hourPart = timePart[0:2]
	//	minPart = timePart[2:4]
	//} else {
	//	log.Fatalf("Too many characters in timePart %s for %s", timePart, rawTsString)
	//}
	//sanitizedTsString := fmt.Sprintf("%s %s %s", hourPart, minPart, splits[1])
	//ts, _ := time.Parse("3 04 PM MST Mon Jan 2 2006", sanitizedTsString)
	ts := adv.PublishedParsed.Truncate(time.Hour)

	// Lat
	lat, _ := strconv.ParseFloat(matchVars[2], 64)
	ns := matchVars[3]
	if strings.ToLower(ns) == "s" {
		lat = -lat
	}

	// Lon
	lon, _ := strconv.ParseFloat(matchVars[4], 64)
	ew := matchVars[5]
	if strings.ToLower(ew) == "w" {
		lon = -lon
	}

	// Name
	if name != "" {
		fmt.Println(matchVars)
	}

	// vMax
	vMaxMph, _ := strconv.ParseFloat(matchVars[6], 64)
	vMaxKts := vMaxMph * mphToKts

	// Bearing
	bearingDeg, _ := strconv.ParseFloat(matchVars[7], 64)

	// Forward Speed
	fSpeedMph, _ := strconv.ParseFloat(matchVars[8], 64)
	fSpeedKts := fSpeedMph * mphToKts

	// Central Pressure
	minCpMb, _ := strconv.ParseFloat(matchVars[9], 64)

	// Discussion
	discussionText := strings.TrimSpace(matchVars[10])

	info = StormFeedInfo{
		Name:            name,
		StormID:         strings.ToLower(stormID),
		AdvNumber:       advNumber,
		Timestamp:       ts.UTC(),
		LatY:            lat,
		LonX:            lon,
		VMaxKts:         vMaxKts,
		BearingDeg:      bearingDeg,
		ForwardSpeedKts: fSpeedKts,
		MinCpMb:         minCpMb,
		Discussion:      discussionText,
		Raw:             adv.Description,
	}
	return info
}

func isGraphics(adv *gofeed.Item) bool {
	return graphicsRegex.MatchString(adv.Title)
}

func parseStormGraphics(adv *gofeed.Item) (name string, links []string) {
	name = graphicsRegex.ReplaceAllString(adv.Title, "$1")
	links = adv.Links
	return name, links
}

func ParseFeed() {
	ParseFeedSave(true)
}
func ParseFeedSave(doSave bool) {
	fp := gofeed.NewParser()

	feed, _ := fp.ParseURL("https://www.nhc.noaa.gov/index-at.xml")

	storms := map[string]StormFeedInfo{}
	links := map[string][]string{}
	sources := map[string][]string{}
	for _, adv := range feed.Items {
		if isPublicAdv(adv) {
			stormInfo := parsePublicAdv(adv)
			storms[stormInfo.Name] = stormInfo
			sources[stormInfo.Name] = append(sources[stormInfo.Name], adv.Link)
		}

		if isGraphics(adv) {
			name, graphics := parseStormGraphics(adv)
			links[name] = graphics
			sources[name] = append(sources[name], adv.Link)
		}

	}

	for name := range storms {
		if l, ok := links[name]; ok {
			info := storms[name]
			info.SetGraphics(l)
			storms[name] = info
		}

		if s, ok := sources[name]; ok {
			info := storms[name]
			info.SetSources(s)
			storms[name] = info
		}
	}

	out, _ := json.MarshalIndent(storms, "", "  ")
	fmt.Println(string(out))

	if doSave {
		for _, storm := range storms {
			//saveToFirestoreDB(storm)
			saveToPostgresDB(storm)
		}
	}
}

func saveToFirestoreDB(info StormFeedInfo) {
	ctx := context.Background()

	projectID, ok := os.LookupEnv("GCP_PROJECT")
	if !ok {
		projectID = "godin-324403"
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	docId := fmt.Sprintf("%s_%d", info.StormID, info.AdvNumber)

	_, err = client.Collection("storms").Doc(info.StormID).Collection("adv").Doc(strconv.Itoa(info.AdvNumber)).Create(ctx, info)
	if err != nil {
		log.Printf("Doc error: %v", err)
	}
	_, err = client.Collection("pending").Doc(docId).Create(ctx, info)

	if err != nil {
		log.Printf("Doc error: %v", err)
	}
	_ = client.Close()
}

type NhcRssWrite struct {
	StormId string `db:"storm_id"`
	AdvNum  int    `db:"adv_num"`
	Raw     string `db:"raw"`
	Parsed  string `db:"parsed"`
}
type NhcRss struct {
	Id        string    `db:"id"`
	StormId   string    `db:"storm_id"`
	AdvNum    int       `db:"adv_num"`
	Raw       string    `db:"raw"`
	Parsed    string    `db:"parsed"`
	Processed bool      `db:"processed"`
	CreatedAt time.Time `db:"created_at"`
}

func (i *StormFeedInfo) toNhcRssTable() (ret NhcRssWrite) {
	parsed, _ := json.Marshal(i)

	return NhcRssWrite{
		StormId: i.StormID,
		AdvNum:  i.AdvNumber,
		Raw:     i.Raw,
		Parsed:  string(parsed),
	}
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func saveToPostgresDB(info StormFeedInfo) {
	fi, _ := os.Open("./sql/db.json")

	var conf DBConfig
	jsonBytes, _ := ioutil.ReadAll(fi)

	_ = json.Unmarshal(jsonBytes, &conf)

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/postgres?sslmode=disable", conf.Username, conf.Password, conf.Host, conf.Port)
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	toWrite := info.toNhcRssTable()
	tx := db.MustBegin()
	res, err := tx.NamedExec(`INSERT INTO odin.nhc_rss (storm_id, adv_num, parsed) VALUES (:storm_id, :adv_num, :parsed)`, &toWrite)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
	_ = tx.Commit()
}
