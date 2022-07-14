/*
	Functions related to adding a new bookmark to Raindrop.io via Alfred

	By Andreas Westerlind in 2021
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	aw "github.com/deanishe/awgo"
)

func select_collection(query string, bookmark_url string, bookmark_title string, firefox_json string, full_collection_paths bool) {
	if firefox_json != "" {
		var firefox_interface map[string]interface{}
		json.Unmarshal([]byte(firefox_json), &firefox_interface)
		bookmark_url = firefox_interface["alfredworkflow"].(map[string]interface{})["variables"].(map[string]interface{})["FF_URL"].(string)
		bookmark_title = firefox_interface["alfredworkflow"].(map[string]interface{})["variables"].(map[string]interface{})["FF_TITLE"].(string)
	}

	// Fix bookmark_url if it has been escaped en extra time, which Arc seems to currently do (they will probably fix that eventually)
	if strings.HasPrefix(bookmark_url, "\"") {
		json.Unmarshal([]byte(bookmark_url), &bookmark_url)
	}

	render_style := "tree"
	if full_collection_paths {
		render_style = "paths"
	}

	// Try to read token, and initiate authentication mekanism if it fails
	token := read_token()
	if token.Error != "" {
		init_auth()
		return
	}

	check_token_lifetime(token)

	if bookmark_title == "" && bookmark_url != "" {
		if _, err := url.ParseRequestURI(bookmark_url); err == nil {
			bookmark_title = get_title(bookmark_url)
			//bookmark_title = "TEST"
		} else {
			bookmark_url = "No browser active"
		}
	}

	if bookmark_url == "No browser active" {
		// Result we didn't get any URL to save, probably because no browser is the frontmost app
		alfred_item := wf.NewItem("There is nothing here to add to Raindrop.io").
			Subtitle("Go to the browser you want to add from, or copy an address first").
			Arg("")
		alfred_item.Alt().
			Subtitle("Go to the browser you want to add from, or copy an address first").
			Arg("")
		return
	}

	bookmark_info := make(map[string]string)
	bookmark_info["collection"] = "-1"
	bookmark_info["title"] = bookmark_title
	bookmark_info["url"] = bookmark_url
	bookmark_json, _ := json.Marshal(bookmark_info)

	// Put alternative to add the new bookmark to Unsorted above the collection list
	alfred_item := wf.NewItem("Add Raindrop.io Bookmark to Unsorted").
		Arg(bookmark_title).
		Subtitle("Or select a collection below").
		Var("bookmark_info", string(bookmark_json)).
		Valid(true)
	alfred_item.Alt().
		Arg(bookmark_title).
		Subtitle("Or select a collection below").
		Var("bookmark_info", string(bookmark_json))
	alfred_item.Cmd().
		Subtitle("Save now, without setting custom title or adding tags").
		Var("bookmark_info", string(bookmark_json)).
		Var("goto", "save now")

	// Get collections
	var raindrop_collections []interface{}
	var raindrop_collections_sublevel []interface{}
	if query != "" {
		raindrop_collections = reverse_interface_array(get_collections(token, false, "trust"))
		raindrop_collections_sublevel = reverse_interface_array(get_collections(token, true, "trust"))
	} else {
		raindrop_collections = reverse_interface_array(get_collections(token, false, "check"))
		raindrop_collections_sublevel = reverse_interface_array(get_collections(token, true, "check"))
	}

	var current_object []string
	render_collections(raindrop_collections, raindrop_collections_sublevel, render_style, "adding", 0, current_object, -1, bookmark_title, bookmark_url)

	// Add Alfred variables for info about the new bookmark
	wf.Var("bookmark_title", bookmark_title)

	// Filter output if search query is entered
	if query != "" {
		wf.Filter(strings.ToLower(query))
	}
}

func set_title(title string) {
	original_title := wf.Config.Get("bookmark_title", "")
	var selection_map map[string]string
	json.Unmarshal([]byte(wf.Config.Get("bookmark_info", "")), &selection_map)
	selection_map["title"] = title
	selection_json, _ := json.Marshal(selection_map)

	alfred_item := wf.NewItem("Save as: "+title).
		Subtitle("Original title: "+original_title).
		Arg().
		Var("bookmark_info", string(selection_json)).
		Valid(true)
	alfred_item.Alt().
		Subtitle("Original title: "+original_title).
		Var("bookmark_info", string(selection_json)).
		Arg()
	alfred_item.Cmd().
		Subtitle("Save now, without adding tags").
		Var("bookmark_info", string(selection_json)).
		Arg().
		Var("goto", "save now")
}

func set_tags(tags string) {
	// Load bookmark_info from file if we cant reach it from the workflow variable
	bookmark_info := wf.Config.Get("bookmark_info", "")
	if bookmark_info == "" {
		bookmark_info_bytes, _ := os.ReadFile(wf.CacheDir() + "/bookmark_info.tmp")
		bookmark_info = string(bookmark_info_bytes)
		os.Remove(wf.CacheDir() + "/bookmark_info.tmp")
	}

	tag_array := strings.Split(tags, ",")
	for i := 0; i < len(tag_array); i++ {
		tag_array[i] = strings.Trim(tag_array[i], " #")
	}

	// Read token and related data from file
	token := read_token()

	var filtered_tags []string

	if tags != "" && tag_array[len(tag_array)-1] != "" {
		// Get tag list from cache
		raindrop_tags := get_tags(token, "trust")

		for _, item_interface := range raindrop_tags {
			item := item_interface.(map[string]interface{})
			if strings.Contains(item["_id"].(string), tag_array[len(tag_array)-1]) {
				filtered_tags = append(filtered_tags, item["_id"].(string))
			}
		}
	} else {
		// Get tag list from Raindrop.io and cache it
		get_tags(token, "check")
	}

	tag_list := ""
	valid_tag_count := 0

	for _, current_tag := range tag_array {
		if current_tag != "" {
			tag_list += "#" + current_tag + " "
			valid_tag_count++
		}
	}

	previous_tags := ""
	pos := strings.LastIndex(tags, ",")
	if pos != -1 {
		previous_tags = tags[0:pos] + ", "
	}

	tag_info := "Save without tags "
	if valid_tag_count == 1 {
		tag_info = "Save with tag "
	}
	if valid_tag_count > 1 {
		tag_info = "Save with tags "
	}

	alfred_item := wf.NewItem(tag_info + tag_list).
		Subtitle("Separate multiple tags with comma: tag1, tag2, tag3").
		Arg(tags).
		Valid(true)
	alfred_item.Alt().
		Subtitle("Separate multiple tags with comma: tag1, tag2, tag3").
		Arg(tags)

	if tags != "" {
		for _, current_tag := range filtered_tags {
			alfred_item := wf.NewItem(current_tag).
				Subtitle("").
				Arg(previous_tags+current_tag+", ").
				Var("goto", "more").
				Valid(true).
				Icon(&aw.Icon{"tag.png", ""})
			alfred_item.Alt().
				Subtitle("").
				Arg(previous_tags+current_tag+", ").
				Var("goto", "more")
			alfred_item.Cmd().
				Subtitle("Add this tag and save").
				Arg(previous_tags + current_tag)
		}
	}

	wf.Var("bookmark_info", bookmark_info)
	os.WriteFile(wf.CacheDir()+"/bookmark_info.tmp", []byte(wf.Config.Get("bookmark_info", "")), 0666)
}

func save_bookmark(tags string) {
	var selection_map map[string]string
	json.Unmarshal([]byte(wf.Config.Get("bookmark_info", "")), &selection_map)

	// Prepare tags
	tag_array := strings.Split(strings.Trim(tags, ", "), ",")
	for i := 0; i < len(tag_array); i++ {
		tag_array[i] = strings.Trim(tag_array[i], " #")
	}

	// Read token and related data from file.
	// We assume that this exists, as it would not be possible to get here from within Alfred otherwise.
	token := read_token()

	// Prepare POST variables
	collection_id, _ := strconv.Atoi(selection_map["collection"])
	post_variables := map[string]interface{}{
		"collection": struct {
			Ref string `json:"$ref"`
			Id  int    `json:"$id"`
		}{"collections", collection_id},
		"link":    selection_map["url"],
		"title":   selection_map["title"],
		"tags":    tag_array,
		"excerpt": get_meta_description(selection_map["url"]),
	}
	post_json, _ := json.Marshal(post_variables)

	request, _ := http.NewRequest("POST", "https://api.raindrop.io/rest/v1/raindrop", bytes.NewBuffer(post_json))
	request.Header.Set("User-Agent", "Alfred (Macintosh; Mac OS X)")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	fmt.Print(selection_map["title"])
}
