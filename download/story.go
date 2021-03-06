package igdl

import (
	"fmt"
	"os"
	"strconv"

	"github.com/siongui/instago"
)

func getStoryItem(item instago.IGItem) {
	if !(item.MediaType == 1 || item.MediaType == 2) {
		fmt.Println("In getStoryItem: not single photo or video!")
		return
	}

	urls, err := item.GetMediaUrls()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(urls) != 1 {
		fmt.Println("In getStoryItem: number of download url != 1")
		return
	}
	url := urls[0]

	filepath := getStoryFilePath(
		item.GetUsername(),
		item.GetUserId(),
		item.GetPostCode(),
		url,
		item.GetTimestamp())

	CreateFilepathDirIfNotExist(filepath)
	// check if file exist
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// file not exists
		printDownloadInfo(item, url, filepath)
		err = Wget(url, filepath)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

// DownloadUserStoryByName downloads unexpired stories (last 24 hours) of the
// given user name.
func (m *IGDownloadManager) DownloadUserStoryByName(username string) {
	id, err := instago.GetUserId(username)
	if err != nil {
		panic(err)
	}

	tray, err := m.apimgr.GetUserStory(id)
	if err != nil {
		panic(err)
	}
	for _, item := range tray.GetItems() {
		getStoryItem(item)
	}
	return
}

// DownloadUserStory downloads unexpired stories (last 24 hours) of the given
// user id.
func (m *IGDownloadManager) DownloadUserStory(userId int64) (err error) {
	tray, err := m.apimgr.GetUserStory(strconv.FormatInt(userId, 10))
	if err != nil {
		return
	}
	for _, item := range tray.GetItems() {
		getStoryItem(item)
	}
	return
}

// DownloadUnreadStory downloads all available stories in IGReelTray.
func DownloadUnreadStory(trays []instago.IGReelTray) {
	for _, tray := range trays {
		//fmt.Println(tray.GetUsername())
		for _, item := range tray.GetItems() {
			getStoryItem(item)
		}
	}
}

func (m *IGDownloadManager) fetchUserStory(userId int64, username string, c chan int) {
	defer func() { c <- 1 }()

	err := m.DownloadUserStory(userId)
	if err != nil {
		fmt.Println("In fetchUserStorie: fail to fetch " + username)
		fmt.Println(err)
	}
}

// DownloadAllStory downloads all unexpired stories of all users in IGReelTray.
func (m *IGDownloadManager) DownloadAllStory(trays []instago.IGReelTray) {
	c := make(chan int)
	numOfStoryUser := 0
	for _, tray := range trays {
		items := tray.GetItems()
		if len(items) == 0 {
			numOfStoryUser++
			go m.fetchUserStory(tray.Id, tray.GetUsername(), c)
		} else {
			for _, item := range items {
				getStoryItem(item)
			}
		}
	}

	// wait all goroutines to finish
	for i := 0; i < numOfStoryUser; i++ {
		<-c
	}
}
