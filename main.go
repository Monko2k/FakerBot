package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	oppai "github.com/flesnuk/oppai5"
	"golang.org/x/net/websocket"
	"gopkg.in/irc.v3"
)

type ConfigData struct {
	TwitchUser	string
	TwitchPass	string
	BanchoUser	string
	BanchoPass	string
	OsuApiKey	string
	GosuPort 	string
}

type GosuData struct {
	Settings struct {
		Folders       struct {
			Skin  string `json:"skin"`
		} `json:"folders"`
	} `json:"settings"`
	Menu struct {
		Bm struct {
			ID           int    `json:"id"`
			Metadata     struct {
				Artist     string `json:"artist"`
				Title      string `json:"title"`
				Difficulty string `json:"difficulty"`
			} `json:"metadata"`
		} `json:"bm"`
	} `json:"menu"`
}

var Config ConfigData

type ApiData []struct {
	BeatmapID           string `json:"beatmap_id"`
	HitLength           string `json:"hit_length"`
	Version             string `json:"version"`
	Artist              string `json:"artist"`
	Title               string `json:"title"`
	Difficultyrating    string `json:"difficultyrating"`
}

func init() {
	readin, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalln("Error: " + err.Error())
	} 
	_ = json.Unmarshal(readin, &Config)
}

func main() {
	requests := make(chan string)
	var IngameData GosuData 
	go Game(&IngameData)
	go Twitch(requests, &IngameData)
	go Bancho(requests)
	
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}

func Game(data *GosuData) {

	origin := "http://localhost/"
	url := "ws://localhost:" + Config.GosuPort + "/ws"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println(err)
	}
	log.Println("Connected to Gosumemory")
	for {
		err = websocket.JSON.Receive(ws, &data)
		if err != nil {
			log.Println(err.Error())
			log.Println("Gosumemory Websocket Disconnected, attempting to reconnect")
		}
	}

	

	
}


func Bancho(out <-chan string) {
	conn, err := net.Dial("tcp", "cho.ppy.sh:6667")
	if err != nil {
		log.Println("Bancho Dial failure", err)
	}

	config := irc.ClientConfig{
		Nick: Config.BanchoUser,
		Pass: Config.BanchoPass,

		Handler: irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
			HandleBancho(c, m, out)
		}),
	}

	client := irc.NewClient(conn, config)
	err = client.Run()
	if err != nil {
		log.Println("Bancho Create failure", err)
	}
	log.Println("Bancho Disconnected, attempting to reconnect")
	
}

func HandleBancho(c *irc.Client, m *irc.Message, out <-chan string) {
	if m.Command == "001" {
		c.Write("JOIN " + Config.BanchoUser)
		log.Println("Connected to Bancho")
		for {
			msg := <- out
			c.WriteMessage(&irc.Message{
				Command: "PRIVMSG",
				Params: []string{
					Config.BanchoUser,
					msg,
				},
			})
		}				
	} 

}

func Twitch(out chan<-string, data *GosuData) {
	conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		log.Println("Twitch Dial failure", err)
	}
	config := irc.ClientConfig{
		Nick: Config.TwitchUser,
		Pass: Config.TwitchPass,

		Handler: irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
			HandleTwitch(c, m, out, data)
		}),
	}

	client := irc.NewClient(conn, config)
	err = client.Run()
	if err != nil {
		log.Println("Twitch Create failure", err)
	}
	log.Println("Bancho Disconnected, attempting to reconnect")

}

func HandleTwitch(c *irc.Client, m *irc.Message, out chan<-string, data *GosuData) {
	if m.Command == "001" {
		c.Write("JOIN #" + strings.ToLower(Config.TwitchUser)) // lol
		log.Println("Connected to Twitch")
	} else if m.Command == "PRIVMSG" && c.FromChannel(m) {
		message := strings.ToLower(m.Params[1])
		urlregex, _ := regexp.Compile(`https:\S+`)
		if urlregex.MatchString(message) {
			beatmap_link := urlregex.FindString(message)
			var is_b_link bool
			var is_s_link bool
			undetermined_link, _ := regexp.Compile(`^https:\/\/osu.ppy.sh\/beatmapsets`)
			if undetermined_link.MatchString(beatmap_link) {
				is_b_link = strings.Contains(beatmap_link, "#osu") 
				is_s_link = !is_b_link
			} else {
				b_link_regex, _ := regexp.Compile(`(^https:\/\/osu.ppy.sh\/b\/)|(^https:\/\/old.ppy.sh\/b\/)|(^https:\/\/osu.ppy.sh\/beatmaps)`)
				s_link_regex, _ := regexp.Compile(`(^https:\/\/osu.ppy.sh\/s\/)|(^https:\/\/old.ppy.sh\/s\/)`)
				is_b_link = b_link_regex.MatchString(beatmap_link)
				is_s_link = s_link_regex.MatchString(beatmap_link)
			}

			if is_b_link || is_s_link {
				
				beatmap_idregex, _ := regexp.Compile(`\d+$`)
				if beatmap_idregex.MatchString(beatmap_link) {
					hd, _ := regexp.Compile(`(?i)(hd)|(hidden)`)
					hr, _ := regexp.Compile(`(?i)(hr)|(hardrock)|(hard rock)`)
					dt, _ := regexp.Compile(`(?i)(dt)|(nc)|(doubletime)|(double time)|(nightcore)|(night core)`)
					ez, _ := regexp.Compile(`(?i)(ez)|(easy)`)
					fl, _ := regexp.Compile(`(?i)(fl)|(flashlight)|(flash light)`)
					ht, _ := regexp.Compile(`(?i)(ht[^t])|(ht$)|(halftime)|(half time)`)


					var mods uint32 = 0
					modstring := ""
					if hd.MatchString(message) {
						modstring += "HD,"
					}
					if hr.MatchString(message) {
						modstring += "HR,"
						mods += (1<<4)
					}
					if dt.MatchString(message) {
						modstring += "DT,"
						mods += (1<<6)
					}
					if ez.MatchString(message) {
						modstring += "EZ,"
						mods += (1<<1)
					}
					if fl.MatchString(message) {
						modstring += "FL,"
					}
					if ht.MatchString(message) {
						modstring += "HT,"
						mods += (1<<8)
					}
					if strings.HasSuffix(modstring, ",") {
						modstring = strings.TrimSuffix(modstring, ",")
						modstring += " "
					}


					url := "https://osu.ppy.sh/api/get_beatmaps?k=" + Config.OsuApiKey + "&b=" + beatmap_idregex.FindString(beatmap_link)
					client := &http.Client {
					}
					req, err := http.NewRequest("GET", url, nil)

					if err != nil {
						log.Println(err)
						return
					}
					res, err := client.Do(req)
					if err != nil {
						log.Println(err)
						return
					}
					defer res.Body.Close()

					body, err := ioutil.ReadAll(res.Body)
					if err != nil {
						log.Println(err)
						return
					}

					var response ApiData

					if err := json.Unmarshal(body, &response); err != nil {
						log.Println(err.Error())
						return
					}
					apiresponse := response[0]
					var sr float64
					if mods > 0 {
						//SHIT WILL JUST NOT DO ANYTHING IF THERE IS AN ERROR LOL!
						//DEAL WITH THIS LATER I CBA
						url := "https://osu.ppy.sh/osu/" + apiresponse.BeatmapID
						client := &http.Client {
						}
						req, err := http.NewRequest("GET", url, nil)

						if err != nil {
							fmt.Println(err)
							return
						}
						res, err := client.Do(req)
						if err != nil {
							fmt.Println(err)
							return
						}
						defer res.Body.Close()

						body, err := ioutil.ReadAll(res.Body)
						if err != nil {
							fmt.Println(err)
							return
						}
						var params oppai.Parameters
						params.Mods = mods
						beatmap := oppai.Parse(bytes.NewReader(body))
						sr = oppai.PPInfo(beatmap, &params).Diff.Total
					} else {	
						sr, _ = strconv.ParseFloat(apiresponse.Difficultyrating, 64)
					}
					truncated_sr := fmt.Sprintf("%.2f", sr)
					hit_length, _ := strconv.Atoi(apiresponse.HitLength)
					if dt.MatchString(message) {
						hit_length = (hit_length * 2)/3
					} else if ht.MatchString(message) {
						hit_length = (hit_length * 3)/2
					}
					formatted_length := fmt.Sprintf("%d:%02d", hit_length/60, hit_length%60)
					metadata_string := fmt.Sprintf("%s - %s [%s] %s(%s\u2605, %s drain length)", apiresponse.Artist, apiresponse.Title, apiresponse.Version, modstring, truncated_sr, formatted_length)
					beatmapmessage := fmt.Sprintf("%s > [https://osu.ppy.sh/b/%s %s]", m.Prefix.User, apiresponse.BeatmapID, metadata_string)
					responsemessage := fmt.Sprintf("%s > %s", m.Prefix.User, metadata_string)
					out<-beatmapmessage
					SendTwitchMessage(c, responsemessage)
				}
			}
		} else {
			log.Println(m.Prefix.User, message)
		}

		if message == "!ping" {
			SendTwitchMessage(c, "pong")
		}
		if message == "!np" {
			npmessage := fmt.Sprintf("%s > %s - %s [%s] (https://osu.ppy.sh/b/%d)", m.Prefix.User, data.Menu.Bm.Metadata.Artist, data.Menu.Bm.Metadata.Title, data.Menu.Bm.Metadata.Difficulty, data.Menu.Bm.ID)
			SendTwitchMessage(c, npmessage)
		}
		if message == "!skin" {
			skinmessage := fmt.Sprintf("%s > Current Skin: %s", m.Prefix.User, data.Settings.Folders.Skin)
			SendTwitchMessage(c, skinmessage)
		}
		
	}
}

func SendTwitchMessage(c *irc.Client, m string) {
	c.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			"#" + Config.TwitchUser,
			m,
		},
	})
}