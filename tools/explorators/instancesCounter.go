package explorators

import (
	"golang.org/x/net/html"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func InstancesUsersCount(b Backend) {
	for {
		instances, err := b.FindInstancesToScan()
		if err == nil {
			wg := new(sync.WaitGroup)
			workers_count := 0
			for _, instance := range instances {
				if workers_count > 100 { // do not consume all pg connectors
					wg.Wait()
					workers_count = 0
				} else {
					workers_count++
				}
				wg.Add(1)
				go func(inst Instance, wg *sync.WaitGroup) {
					users_count, err := getInstanceUsersCount(inst.Domain)
					if err == nil {
						inst.UsersCount = uint(users_count)
						inst.LastCount = time.Now()
						inst.CountFailed = false
					} else {
						inst.CountFailed = true
					}
					b.SaveInstance(inst)
					wg.Done()
				}(instance, wg)
			}
		}
		time.Sleep(3 * time.Hour)
	}
}

func getInstanceUsersCount(instance string) (int, error) {
	client := http.Client{
		Timeout: time.Second * 5,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// request has been redirected, don't try to get a user_count from the wrong hostname
			return http.ErrAbortHandler
		},
	}
	resp, err := client.Get("https://" + instance + "/about/more")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	moreDOM := html.NewTokenizer(resp.Body)
	for {
		// must find <div class='information-board'><div class='section'><span>xxxxx</span><strong>xxxx</strong>
		// 			                                            users count is here ----^
		tokenType := moreDOM.Next()
		switch tokenType {
		case html.ErrorToken:
			return 0, moreDOM.Err()
		case html.StartTagToken:
			t := moreDOM.Token()
			if t.Data == "div" {
				for _, attr := range t.Attr {
					if attr.Key == "class" {
						if attr.Val == "information-board" {
							for {
								moreDOM.Next()
								infoToken := moreDOM.Token()
								if infoToken.Data == "strong" {
									moreDOM.Next()
									user_count := string(moreDOM.Text())
									user_count = strings.Replace(user_count, ",", "", -1)
									user_count = strings.Replace(user_count, ".", "", -1)
									user_count = strings.Replace(user_count, " ", "", -1)
									return strconv.Atoi(user_count)
								}
							}
						}
					}
				}
			}

		}
	}
}
