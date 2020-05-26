package main

import (
	"log"
	"time"
	"os"
	"bufio"
	"errors"
	"strconv"
	"github.com/caarlos0/env/v6"
	"encoding/json"
	"net/http"
    "github.com/dghubble/go-twitter/twitter"
    "github.com/dghubble/oauth1"
)

type Config struct {
	TwitterToken string `env:"T_CONSUMER_TOKEN,required"`
	TwitterSecret string `env:"T_CONSUMER_SECRET,required"`
	TokenBeastLocation string `env:"BEAST_LOCATION,required"`
	TokenBeastSecret string `env:"BEAST_SECRET,required"`
}


func getConfig() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return cfg
}


func getTwitterClient(twitterToken, twitterSecret, userToken, userSecret string) *twitter.Client {
	config := oauth1.NewConfig(twitterToken, twitterSecret)
	token := oauth1.NewToken(userToken, userSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient)
}


func ParseRateLimiting(resp *http.Response) (int, time.Duration) {
	remaining, _ := strconv.Atoi(resp.Header["X-Rate-Limit-Remaining"][0])
	reset, _ := strconv.Atoi(resp.Header["X-Rate-Limit-Reset"][0])
	untilReset := reset - int(time.Now().Unix())
	return remaining, time.Duration(untilReset) * time.Second
}

func defaultHandler(err error) (time.Duration, bool, bool) {
	sleeping := 60 * time.Second
	log.Printf("Twitter Error. Will retry. Sleeping %v seconds. Error: \n%v\n", sleeping, err)
	return sleeping, false, false
}

func ignoreHandler(err error) (time.Duration, bool, bool){
	log.Print(err)
	return 0, false, true
}

func HandleErrors(err error, httpResponse *http.Response, errs chan error) (time.Duration, bool, bool) {
	switch err.(type) {
	case twitter.APIError:

		// could use err.Errors[0].Code, but this seems simpler for now
		switch httpResponse.StatusCode {

		// Twitter rate limits, so sleep until limit resets
		case 429:
			_, reset := ParseRateLimiting(httpResponse)
			log.Printf("Sleeping: %v\n", reset)

			sleeping := reset + time.Second
			return sleeping, true, false

		case 500:
			return defaultHandler(err)
		case 502:
			return defaultHandler(err)
		case 503:
			return defaultHandler(err)
		case 504:
			return defaultHandler(err)

		default:
			errs <- err
			return 0, false, false // won't return
		}

	default:
		return defaultHandler(err)
	}
}


func Sleeper(err error, httpResponse *http.Response, retry func(), errs chan error) {
	sleeping, rateLimit, ignore := HandleErrors(err, httpResponse, errs)

	if ignore {
		return
	}

	if rateLimit {
		retry()
		time.Sleep(sleeping)
		return
	}

	time.Sleep(sleeping)
	retry()
	return
}

func Bool(b bool) *bool { return &b }

type AgriusClient struct {
	client *twitter.Client
}

func newAgriusClient(twitterToken, twitterSecret, userToken, userSecret string) *AgriusClient {
	return &AgriusClient{
		client: getTwitterClient(twitterToken, twitterSecret, userToken, userSecret),
	}
}


func OldestTweet(tweets []twitter.Tweet) (int64, error) {
	if len(tweets) == 0 {
		return 0, errors.New("no tweets")
	}

	mx := tweets[0].ID
	for _, tweet := range tweets {
		if tweet.ID < mx {
			mx = tweet.ID
		}
	}

	return mx, nil
}



type UserTimelineWork struct {
	params *twitter.UserTimelineParams
	retries int
}


func (work *UserTimelineWork) Do(client *twitter.Client, inchan chan TwitterTweetWork, outchan chan *twitter.Tweet, errs chan error) {

	log.Printf("Making call with max %d", work.params.MaxID)

	tweets, httpResponse, err := client.Timelines.UserTimeline(work.params)

	if err != nil {
		Sleeper(err, httpResponse, func() { inchan <- work }, errs)
		return
	}

	for _, tweet := range tweets {
		outchan <- &tweet
	}

	if len(tweets) == 0 {
		log.Printf("NO MORE TWEETS: %d", work.params.MaxID)
		// no more tweets to get
		if work.retries > 10 {
			close(inchan)
			return
		}

		time.Sleep(2*time.Second)
		work.retries += 1
		inchan <- work
		return
	}

	time.Sleep(2*time.Second)

	mx, _ := OldestTweet(tweets)
	work.params.MaxID = mx
	inchan <- work
}


type UserLookupWork struct {
	params *twitter.UserLookupParams
}


func (work *UserLookupWork) Do(client *twitter.Client, inchan chan TwitterUserWork, outchan chan *twitter.User, errs chan error) {

	users, httpResponse, err := client.Users.Lookup(work.params)

	if err != nil {
		Sleeper(err, httpResponse, func() { inchan <- work }, errs)
	}

	for _, user := range users {
		outchan <- &user
	}

	close(inchan)
}


type UserShowWork struct {
	params *twitter.UserShowParams
}

type AgriusUserStatus struct {
	*twitter.User
	Suspended bool `json:"suspended,omitempty"`
	Missing bool `json:"missing,omitempty"`
}


func (work *UserShowWork) Do(client *twitter.Client, inchan chan AgriusUserStatusWork, outchan chan *AgriusUserStatus, errs chan error) {

	user, httpResponse, err := client.Users.Show(work.params)
	au := &AgriusUserStatus{User: user}

	if err != nil {
		if e, ok := err.(twitter.APIError); ok {
			switch e.Errors[0].Code {
			case 50:
				au.Missing = true
				au.ScreenName = work.params.ScreenName
			case 63:
				au.Suspended = true
				au.ScreenName = work.params.ScreenName
			default:
				Sleeper(err, httpResponse, func() { inchan <- work }, errs)
				return
			}
		}
	}

	outchan <- au
	close(inchan)
}


func UserShow(username string) *UserShowWork {
	return &UserShowWork{
		params: &twitter.UserShowParams{
			ScreenName: username,
			IncludeEntities: Bool(true),
		},
	}
}


func UserLookup(usernames []string) *UserLookupWork {
	return &UserLookupWork{
		params: &twitter.UserLookupParams{
			ScreenName: usernames,
			IncludeEntities: Bool(true),
		},
	}
}

func UserTimeline(name string) *UserTimelineWork {

	return &UserTimelineWork{
		params: &twitter.UserTimelineParams{
			ScreenName: name,
			Count: 200,
			IncludeRetweets: Bool(true),
		},
	}
}


func monitor(errs <-chan error) {
	e := <- errs
	log.Fatalf("Cage failed with error: %v", e)
}


func check(e error) {
    if e != nil {
        panic(e)
    }
}


// func main() {
// 	cnf := getConfig()

// 	inch, outch, errs := TwitterTweetMasterPool(cnf, 1)
// 	go monitor(errs)
// 	inch <- UserTimeline("realDonaldTrump")
// 	close(inch)

// 	i := 0

//     f, err := os.Create("baz.json")
//     check(err)

// 	w := bufio.NewWriter(f)

// 	for tw := range outch {
// 		t, _ := tw.CreatedAtTime()
// 		log.Print(t)

// 		i += 1
// 		s, err := json.Marshal(&tw)
// 		if err == nil {
// 			w.Write(s)
// 			w.WriteString("\n")
// 		} else {
// 			log.Print(tw)
// 		}

// 	}

// 	w.Flush()

// 	log.Print(i)
// }

func ReadUsersFile(fi string) chan []string {
	f, _ := os.Open(fi)
	scanner := bufio.NewScanner(f)

	ch := make(chan []string)
	go func() {
		defer close(ch)

		users := make([]string, 0)

		for scanner.Scan() {
			user := scanner.Text()
			users = append(users, user)

			if len(users) == 100 {
				ch <- users
				users = make([]string, 0)
			}
		}
	}()

	return ch
}


// func main() {
// 	cnf := getConfig()

// 	inch, outch, errs := TwitterUserMasterPool(cnf, 20)
// 	go monitor(errs)

// 	fi := ReadUsersFile("users")
// 	go func() {
// 		defer close(inch)
// 		for users := range fi {
// 			work := UserLookup(users)
// 			inch <- work
// 		}
// 	}()

// 	f, err := os.Create("users.json")
//     check(err)

// 	w := bufio.NewWriter(f)

// 	for user := range outch {
// 		// log.Print(user)
// 		s, err := json.Marshal(&user)
// 		if err == nil {
// 			w.Write(s)
// 			w.WriteString("\n")
// 		} else {
// 			log.Print(user)
// 		}
// 	}

// 	w.Flush()
// }



func main() {
	cnf := getConfig()

	inch, outch, errs := AgriusUserStatusMasterPool(cnf, 20)
	go monitor(errs)

	fi := ReadUsersFile("missing")
	go func() {
		defer close(inch)
		for users := range fi {
			for _, user := range users {
				work := UserShow(user)
				inch <- work
			}
		}
	}()

	f, err := os.Create("users-missing.json")
    check(err)

	w := bufio.NewWriter(f)

	for user := range outch {
		// log.Print(user)
		s, err := json.Marshal(&user)
		if err == nil {
			w.Write(s)
			w.WriteString("\n")
		} else {
			log.Print(user)
		}
	}

	w.Flush()
}
