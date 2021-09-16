package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type PatchList struct {
	Patches []Patch `json:"patches"`
	Success bool    `json:"success"`
}
type Patch struct {
	PatchNumber        string `json:"patch_number"`
	PatchName          string `json:"patch_name"`
	PatchTimestamp     int    `json:"patch_timestamp"`
	PatchWebsite       string `json:"patch_website,omitempty"`
	PatchWebsiteAnchor string `json:"patch_website_anchor,omitempty"`
}

func LoadPatches(pl *PatchList, url string) {
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		log.Println(err)
		return
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	err = json.Unmarshal(body, pl)
	if err != nil {
		log.Fatal("error: ", err)
	}
}

func Notify(patch Patch, url string) {

	type Payload struct {
		Text        string `json:"text"`
		Format      string `json:"format"`
		DisplayName string `json:"displayName"`
		AvatarURL   string `json:"avatar_url"`
	}
	var msg string
	if patch.PatchNumber == "" {
		msg = "DotaPatchBot online"
	} else {
		log.Printf("Patch %v found, notifying\n", patch.PatchNumber)

		if patch.PatchWebsite != "" {
			msg = fmt.Sprintf("Patch %v released, see website:\nhttps://www.dota2.com/%v", patch.PatchNumber, patch.PatchWebsite)
		} else {
			msg = fmt.Sprintf("Patch %v Released", patch.PatchNumber)
		}
	}
	if url != "" {
		data := struct {
			Text        string `json:"text"`
			Format      string `json:"format"`
			DisplayName string `json:"displayName"`
			AvatarURL   string `json:"avatar_url"`
		}{
			msg,
			"plain",
			"DotaPatchBot",
			"https://i.pinimg.com/originals/8a/8b/50/8a8b50da2bc4afa933718061fe291520.jpg",
		}

		payloadBytes, err := json.Marshal(data)
		if err != nil {
			// handle err
		}
		body := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			// handle err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// handle err
		}
		defer resp.Body.Close()
	}
}
func main() {
	hookPtr := flag.String("hook", "", "the webhook to notify")
	sourcePtr := flag.String("url", "https://www.dota2.com/datafeed/patchnoteslist", "url to fetch patchnotes from")
	patch_list := PatchList{}
	var current_patch Patch

	flag.Parse()
	LoadPatches(&patch_list, *sourcePtr)

	current_patch = patch_list.Patches[len(patch_list.Patches)-1]
	log.Println("Started bot, loaded patch list")
	log.Printf("Starting on patch %v", current_patch.PatchNumber)
	if *hookPtr == "" {
		log.Println("No webhook set, not notifying")
	} else {
		log.Printf("Notifying %v", *hookPtr)
	}
	Notify(Patch{}, *hookPtr)

	for {
		time.Sleep(30 * time.Second)
		LoadPatches(&patch_list, *sourcePtr)
		newest_patch := patch_list.Patches[len(patch_list.Patches)-1]
		if newest_patch.PatchNumber != current_patch.PatchNumber {
			current_patch = newest_patch
			if *hookPtr != "" {
				Notify(current_patch, *hookPtr)
			} else {
				log.Printf("Patch %v found, no webhook\n", current_patch.PatchNumber)
			}
		}
	}

}
