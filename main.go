package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type patchList struct {
	Patches []patch `json:"patches"`
	Success bool    `json:"success"`
}
type patch struct {
	PatchNumber        string `json:"patch_number"`
	PatchName          string `json:"patch_name"`
	PatchTimestamp     int    `json:"patch_timestamp"`
	PatchWebsite       string `json:"patch_website,omitempty"`
	PatchWebsiteAnchor string `json:"patch_website_anchor,omitempty"`
}

func loadPatches(pl *patchList, url string) {
	res, err := http.Get(url)
	if err != nil {
		log.Printf("Got %v, retrying in 5s", err)
		time.Sleep(5 * time.Second)
		res, err = http.Get(url)
		if err != nil {
			log.Printf("Got %v, not trying again", err)
			return
		}
	}
	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	err = json.Unmarshal(body, pl)
	if err != nil {
		log.Fatal("error: ", err)
	}
}

func notify(ptch patch, url string) {
	type Payload struct {
		Text        string `json:"text"`
		Format      string `json:"format"`
		DisplayName string `json:"displayName"`
		AvatarURL   string `json:"avatar_url"`
	}
	var msg string
	if ptch.PatchNumber == "" {
		msg = "DotaPatchBot online"
	} else {
		log.Printf("Patch %v found, notifying\n", ptch.PatchNumber)

		if ptch.PatchWebsite != "" {
			msg = fmt.Sprintf("Patch %v released, see website:\nhttps://www.dota2.com/%v", ptch.PatchNumber, ptch.PatchWebsite)
		} else {
			msg = fmt.Sprintf("Patch %v Released", ptch.PatchNumber)
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
			log.Fatalf("Can't create webhook notice; failing since that means drastic data format change")
		}
		body := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			log.Fatalf("Either POST was mispelled or couldn't create a context")
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("error posting to webhook")
		}
		if resp.Body != nil {
			resp.Body.Close()
		}
	}
}

func main() {
	hookPtr := flag.String("hook", "", "the webhook to notify")
	sourcePtr := flag.String("url", "https://www.dota2.com/datafeed/patchnoteslist", "url to fetch patchnotes from")
	patchlist := patchList{}
	var currentPatch patch

	flag.Parse()
	if *hookPtr == "" {
		if val, ok := os.LookupEnv("DOTA_WEBHOOK"); ok {
			*hookPtr = val
		}
	}
	if *sourcePtr == "" {
		if val, ok := os.LookupEnv("DOTA_PATCH_SOURCE"); ok {
			*sourcePtr = val
		}
	}
	loadPatches(&patchlist, *sourcePtr)

	if len(patchlist.Patches) <= 0 {
		log.Fatal("error getting initial patch list")
	}

	currentPatch = patchlist.Patches[len(patchlist.Patches)-1]
	log.Println("Started bot, loaded patch list")
	log.Printf("Starting on patch %v", currentPatch.PatchNumber)
	if *hookPtr == "" {
		log.Println("No webhook set, not notifying")
	} else {
		log.Printf("Notifying %v", *hookPtr)
	}
	notify(patch{}, *hookPtr)

	for {
		time.Sleep(30 * time.Second)
		loadPatches(&patchlist, *sourcePtr)
		if len(patchlist.Patches) > 0 {
			newestPatch := patchlist.Patches[len(patchlist.Patches)-1]
			if newestPatch.PatchNumber != currentPatch.PatchNumber {
				currentPatch = newestPatch
				if *hookPtr != "" {
					notify(currentPatch, *hookPtr)
				} else {
					log.Printf("Patch %v found, no webhook\n", currentPatch.PatchNumber)
				}
			}
		}
	}
}
