/*
	Functions for local caching and searching Raindrop.io bookmarks in Alfred

	By Andreas Westerlind, 2025
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
)

// Function for fetching and caching all bookmarks from Raindrop.io
func get_all_bookmarks(token RaindropToken, caching string) []interface{} {
	// If caching == "check": Redownload bookmarks only if cache is older than the configured refresh interval
	// If caching == "trust": Trust the bookmarks cache to be good enough and use what is cached without checking its age (only download if no cache exists yet)
	// If caching == "fetch": Always redownload bookmarks without checking the age of the cache

	var bookmarks []interface{}
	var cache_base map[string]interface{}
	var cache_filename string = wf.CacheDir() + "/bookmarks.json"

	// Check if cache files exist
	var bookmarks_cache_exists bool = false
	var collections_cache_exists bool = false
	var collections_sublevel_cache_exists bool = false
	var tags_cache_exists bool = false

	if _, err := os.Stat(wf.CacheDir() + "/bookmarks.json"); err == nil {
		bookmarks_cache_exists = true
	}
	if _, err := os.Stat(wf.CacheDir() + "/collections.json"); err == nil {
		collections_cache_exists = true
	}
	if _, err := os.Stat(wf.CacheDir() + "/collections_sublevel.json"); err == nil {
		collections_sublevel_cache_exists = true
	}
	if _, err := os.Stat(wf.CacheDir() + "/tags.json"); err == nil {
		tags_cache_exists = true
	}

	// If any cache doesn't exist and caching is set to "trust", switch to "fetch"
	if caching == "trust" && (!bookmarks_cache_exists || !collections_cache_exists || !collections_sublevel_cache_exists || !tags_cache_exists) {
		caching = "fetch"
	}

	// Get cache refresh interval from config (default: 24 hours)
	refresh_interval_str := wf.Config.Get("local_cache_refresh_interval", "24")
	refresh_interval, err := strconv.ParseFloat(refresh_interval_str, 64)
	if err != nil {
		refresh_interval = 24 // Default to 24 hours if parsing fails
	}

	// Check if cache file exists
	if cache_file_stat, err := os.Stat(cache_filename); err == nil {
		// Use cache if caching is set to "trust", or if it's not older than the configured refresh interval and caching is set to "check"
		if caching == "trust" || (time.Since(cache_file_stat.ModTime()).Hours() < refresh_interval && caching == "check") {
			// Read stored cached bookmarks
			cache_file, _ := os.ReadFile(cache_filename)
			json.Unmarshal(cache_file, &cache_base)
			if cache_base["items"] != nil && cache_base["items"].([]interface{}) != nil {
				bookmarks = cache_base["items"].([]interface{})
				return bookmarks
			}
		}
	}

	// Cache doesn't exist, is too old, or force refresh is enabled
	// Fetch all bookmarks from Raindrop.io
	all_bookmarks := []interface{}{}
	page := 0
	perPage := 50 // The Raindrop.io API seems to limit results to 50 per page independent of this value

	for {
		// Query Raindrop.io for a page of bookmarks
		client := &http.Client{}
		params := url.Values{
			"perpage": []string{fmt.Sprint(perPage)},
			"page":    []string{fmt.Sprint(page)},
		}
		u := &url.URL{
			Scheme:   "https",
			Host:     "api.raindrop.io",
			Path:     "/rest/v1/raindrops/0",
			RawQuery: params.Encode(),
		}
		request, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return bookmarks
		}
		request.Header.Set("User-Agent", "Alfred (Macintosh; Mac OS X)")
		request.Header.Set("Authorization", "Bearer "+token.AccessToken)
		response, err := client.Do(request)
		if err != nil {
			return bookmarks
		}
		response_body, err := io.ReadAll(response.Body)
		if err != nil {
			return bookmarks
		}

		var result map[string]interface{}
		json.Unmarshal(response_body, &result)

		// Check if we got valid results
		if result["items"] == nil || len(result["items"].([]interface{})) == 0 {
			break // No more bookmarks to fetch
		}

		// Add this page's bookmarks to our cache collection
		page_bookmarks := result["items"].([]interface{})
		all_bookmarks = append(all_bookmarks, page_bookmarks...)

		// If we got fewer bookmarks than requested, we've reached the end
		if len(page_bookmarks) < perPage {
			break
		}

		// Move to the next page
		page++
	}

	// Create a result object with the same structure as the API response
	result := map[string]interface{}{
		"items": all_bookmarks,
	}

	// Write to cache file
	result_json, _ := json.Marshal(result)
	os.WriteFile(cache_filename, result_json, 0666)

	// If we've updated the bookmarks cache, also update tags and collections
	if caching == "fetch" {
		get_tags(token, "fetch")
		get_collections(token, false, "fetch")
		get_collections(token, true, "fetch")
	}

	return all_bookmarks
}

// Function for searching the local bookmark cache
func local_search(query string, token RaindropToken, collection int, tag string, descr_in_list bool, favs_first bool) {
	// Fetch all bookmarks from cache (or from API if no cache exists at all)
	bookmarks := get_all_bookmarks(token, "trust")

	// If we got no bookmarks, show a message and return
	if len(bookmarks) == 0 {
		wf.NewItem("No bookmarks found in cache").
			Subtitle("Try refreshing the cache or check your Raindrop.io account").
			Valid(false)
		return
	}

	// Get collection list from cache
	raindrop_collections := reverse_interface_array(get_collections(token, false, "trust"))
	raindrop_collections_sublevel := reverse_interface_array(get_collections(token, true, "trust"))

	// If no query and not searching in a collection or by tag, show default options
	if query == "" && collection == 0 && tag == "" {
		alfred_item := wf.NewItem("Search your Raindrop.io bookmarks").
			Arg("https://app.raindrop.io/").
			Var("goto", "open").
			Subtitle("Or press enter to open Raindrop.io").
			Valid(true)
		alfred_item.Alt().
			Arg("https://app.raindrop.io/").
			Var("goto", "open").
			Subtitle("Or press enter to open Raindrop.io")
		alfred_item2 := wf.NewItem("Browse your Raindrop.io collections").
			Var("goto", "local_browse").
			Subtitle("").
			Valid(true).
			Icon(&aw.Icon{Value: "folder.png", Type: ""})
		alfred_item2.Alt().
			Var("goto", "local_browse").
			Subtitle("")
	}

	// Filter bookmarks by collection if specified
	if collection != 0 {
		filtered_bookmarks := []interface{}{}
		for _, bookmark := range bookmarks {
			bookmark_map := bookmark.(map[string]interface{})
			if bookmark_map["collection"] != nil {
				collection_map := bookmark_map["collection"].(map[string]interface{})
				if collection_map["$id"] != nil && int(collection_map["$id"].(float64)) == collection {
					filtered_bookmarks = append(filtered_bookmarks, bookmark)
				}
			}
		}
		bookmarks = filtered_bookmarks
	}

	// Filter bookmarks by tag if specified
	if tag != "" {
		filtered_bookmarks := []interface{}{}
		for _, bookmark := range bookmarks {
			bookmark_map := bookmark.(map[string]interface{})
			if bookmark_map["tags"] != nil {
				tags := bookmark_map["tags"].([]interface{})
				for _, t := range tags {
					if strings.ToLower(t.(string)) == strings.ToLower(tag) {
						filtered_bookmarks = append(filtered_bookmarks, bookmark)
						break
					}
				}
			}
		}
		bookmarks = filtered_bookmarks
	}

	// Filter bookmarks by query if specified
	if query != "" {
		filtered_bookmarks := []interface{}{}
		query_lower := strings.ToLower(query)
		for _, bookmark := range bookmarks {
			bookmark_map := bookmark.(map[string]interface{})

			// Check title
			title := ""
			if bookmark_map["title"] != nil {
				title = strings.ToLower(bookmark_map["title"].(string))
			}

			// Check excerpt
			excerpt := ""
			if bookmark_map["excerpt"] != nil {
				excerpt = strings.ToLower(bookmark_map["excerpt"].(string))
			}

			// Check URL
			link := ""
			if bookmark_map["link"] != nil {
				link = strings.ToLower(bookmark_map["link"].(string))
			}

			// Check tags
			tags_str := ""
			if bookmark_map["tags"] != nil {
				tags := bookmark_map["tags"].([]interface{})
				for _, t := range tags {
					tags_str += strings.ToLower(t.(string)) + " "
				}
			}

			// If any field contains the query, add the bookmark to results
			if strings.Contains(title, query_lower) ||
				strings.Contains(excerpt, query_lower) ||
				strings.Contains(link, query_lower) ||
				strings.Contains(tags_str, query_lower) {
				filtered_bookmarks = append(filtered_bookmarks, bookmark)
			}
		}
		bookmarks = filtered_bookmarks
	}

	var current_object []string
	collection_names := collection_paths(raindrop_collections, raindrop_collections_sublevel, make(map[int]string), 0, current_object, -1)

	var render_favourites string = "all"

	// Prepare favourites for being viewed in Alfred (if favourites_first is enabled)
	if favs_first {
		render_results(bookmarks, "only", collection_names, descr_in_list)
		render_favourites = "none"
	}

	// Prepare the rest of the results (or all results if favourites_first is disabled) for being viewed in Alfred
	render_results(bookmarks, render_favourites, collection_names, descr_in_list)

	// If no results after filtering, show a message
	if len(bookmarks) == 0 {
		wf.NewItem("No matching bookmarks found").
			Subtitle("Try a different search query").
			Valid(false)
	}

	// Always show collections and tags at the bottom when doing a local cache search
	if collection == 0 && tag == "" {
		// Render collections
		var current_object []string
		render_collections(raindrop_collections, raindrop_collections_sublevel, "paths", "searching", 0, current_object, -1, "", "", "local")

		// Get tag list from cache
		raindrop_tags := get_tags(token, "trust")

		// Render tags
		for _, item_interface := range raindrop_tags {
			item := item_interface.(map[string]interface{})
			alfred_item := wf.NewItem(item["_id"].(string)).
				Var("current_tag", item["_id"].(string)).
				Var("goto", "local_tag").
				Valid(true).
				Icon(&aw.Icon{Value: "tag.png", Type: ""})
			alfred_item.Alt().
				Var("current_tag", item["_id"].(string)).
				Var("goto", "local_tag").
				Subtitle("")
		}
	}
}

// Main function for handling local search command from Alfred
func local_search_command(variant string, query string, collection_json string, tag string, from string, descr_in_list bool, favs_first bool) {
	// Try to read token, and initiate authentication mechanism if it fails
	token := read_token()
	if token.Error != "" {
		init_auth()
		return
	}

	// Check if token has expired
	time_location, _ := time.LoadLocation("UTC")
	time_format := "2006-01-02 15:04:05"
	token_time, err := time.Parse(time_format, token.CreationTime)
	if err != nil || time.Now().In(time_location).Sub(token_time).Milliseconds() > int64(token.Expires) {
		init_auth()
		return
	}

	// Check if token needs to be refreshed
	check_token_lifetime(token)

	var collection_search bool = false
	var collection_search_id int = 0
	var collection_search_name string
	var collection_search_icon string
	var collection_from_list bool = false
	if variant == "collection" {
		collection_search = true
		var collection_info map[string]string
		json.Unmarshal([]byte(collection_json), &collection_info)
		collection_search_name = collection_info["name"]
		collection_search_id, _ = strconv.Atoi(collection_info["id"])
		collection_search_icon = collection_info["icon"]
		if from == "collections" {
			collection_from_list = true
		}
	}

	var tag_search bool = false
	if variant == "tag" {
		tag_search = true
	}

	if collection_search {
		if collection_from_list {
			// We are browsing a collection, and came here from the collection browser
			alfred_item := wf.NewItem("Bookmarks in "+collection_search_name).
				Var("goto", "browse").
				Subtitle("⬅︎ Go back to collection browser").
				Valid(true).
				Icon(&aw.Icon{Value: collection_search_icon, Type: ""})
			alfred_item.Alt().
				Var("goto", "back").
				Subtitle("⬅︎ Go back to collection browser")
		} else {
			// We are browsing a collection, and came here from the main bookmark search
			alfred_item := wf.NewItem("Bookmarks in "+collection_search_name).
				Var("goto", "back").
				Subtitle("⬅︎ Go back to search all bookmarks").
				Valid(true).
				Icon(&aw.Icon{Value: collection_search_icon, Type: ""})
			alfred_item.Alt().
				Var("goto", "back").
				Subtitle("⬅︎ Go back to search all bookmarks")
		}
	}

	if tag_search {
		// We are browsing bookmarks with a specific tag
		alfred_item := wf.NewItem("Bookmarks tagged with #"+tag).
			Var("goto", "back").
			Subtitle("⬅︎ Go back to search all bookmarks").
			Valid(true).
			Icon(&aw.Icon{Value: "tag.png", Type: ""})
		alfred_item.Alt().
			Var("goto", "back").
			Subtitle("⬅︎ Go back to search all bookmarks")
	}

	// Search the local cache
	local_search(query, token, collection_search_id, tag, descr_in_list, favs_first)

	// Check if the cache needs to be refreshed and spawn a background process if needed
	check_and_refresh_cache()
}

// Function for browsing collections in the local cache
func local_browse(query string, full_collection_paths bool) {
	token := read_token()

	alfred_item := wf.NewItem("Raindrop.io Bookmark Collections").
		Var("goto", "back").
		Subtitle("⬅︎ Go back to search all bookmarks").
		Valid(true).
		Icon(&aw.Icon{Value: "icon.png", Type: ""})
	alfred_item.Alt().
		Var("goto", "back").
		Subtitle("⬅︎ Go back to search all bookmarks")
	alfred_item2 := wf.NewItem("Unsorted").
		Var("collection_info", "{\"icon\":\"folder.png\",\"id\":\"-1\",\"name\":\"Unsorted\"}").
		Var("goto", "local_collection").
		Subtitle("").
		Valid(true).
		Icon(&aw.Icon{Value: "folder.png", Type: ""})
	alfred_item2.Alt().
		Var("collection_info", "{\"icon\":\"folder.png\",\"id\":\"-1\",\"name\":\"Unsorted\"}").
		Var("goto", "local_collection").
		Subtitle("")

	render_style := "tree"
	if full_collection_paths {
		render_style = "paths"
	}

	var raindrop_collections []interface{}
	var raindrop_collections_sublevel []interface{}

	// Get collection list
	raindrop_collections = reverse_interface_array(get_collections(token, false, "trust"))
	raindrop_collections_sublevel = reverse_interface_array(get_collections(token, true, "trust"))

	// Render collections
	var current_object []string
	render_collections(raindrop_collections, raindrop_collections_sublevel, render_style, "searching", 0, current_object, -1, "", "", "local")

	if query != "" {
		wf.Filter(query)
	}

	// Check if the cache needs to be refreshed and spawn a background process if needed
	check_and_refresh_cache()
}

// Function to check if the cache needs to be refreshed
func should_refresh_cache() bool {
	var cache_filename string = wf.CacheDir() + "/bookmarks.json"

	// Get cache refresh interval from config (default: 24 hours)
	refresh_interval_str := wf.Config.Get("local_cache_refresh_interval", "24")
	refresh_interval, err := strconv.ParseFloat(refresh_interval_str, 64)
	if err != nil {
		refresh_interval = 24 // Default to 24 hours if parsing fails
	}

	// Check if cache file exists and if it's older than the refresh interval
	if cache_file_stat, err := os.Stat(cache_filename); err == nil {
		if time.Since(cache_file_stat.ModTime()).Hours() >= refresh_interval {
			return true
		}
	} else {
		// Cache file doesn't exist
		return true
	}

	return false
}

// Function to spawn a background process to refresh the cache
func spawn_background_refresh() {
	// Update the timestamp before triggering the refresh
	update_background_refresh_timestamp()

	// Start a background process to refresh the cache
	cmd := exec.Command("sh", "-c", fmt.Sprintf("nohup ./raindrop_alfred background_refresh > /dev/null 2>&1 & disown"))
	cmd.Start()
}

// Function to handle background refresh of the cache
func background_refresh_cache() {
	// Try to read token
	token := read_token()
	if token.Error != "" {
		return // Can't authenticate in the background
	}

	// Check if token has expired
	time_location, _ := time.LoadLocation("UTC")
	time_format := "2006-01-02 15:04:05"
	token_time, err := time.Parse(time_format, token.CreationTime)
	if err != nil || time.Now().In(time_location).Sub(token_time).Milliseconds() > int64(token.Expires) {
		return // Can't authenticate in the background
	}

	// Check token lifetime and refresh if needed
	check_token_lifetime(token)

	// Force refresh the caches.
	// Unfortunately, there is no way of only getting bookmarks that where changed after the last refresh.
	// New bookmarks would be possible to get this way, but then we would not get changes to existing ones.
	// For that reason, we are simply refetching all bookmarks, with a default to only do it once per day to not overload the Raindrop.io API.
	// Note: get_all_bookmarks will also refresh tags and collections when called with "fetch"
	get_all_bookmarks(token, "fetch")
}

// Function to check if a background refresh was triggered recently (within the last minute)
func was_background_refresh_triggered_recently() bool {
	var timestamp_filename string = wf.CacheDir() + "/background_refresh_timestamp.txt"

	// Check if timestamp file exists
	if timestamp_file_stat, err := os.Stat(timestamp_filename); err == nil {
		// Check if the timestamp is less than 1 minute old
		if time.Since(timestamp_file_stat.ModTime()).Seconds() < 60 {
			return true
		}
	}

	return false
}

// Function to update the background refresh timestamp
func update_background_refresh_timestamp() {
	var timestamp_filename string = wf.CacheDir() + "/background_refresh_timestamp.txt"

	// Create or update the timestamp file
	os.WriteFile(timestamp_filename, []byte(time.Now().String()), 0666)
}

// Function to check if the cache needs to be refreshed and spawn a background process if needed
func check_and_refresh_cache() {
	// Check if the cache needs to be refreshed
	if should_refresh_cache() {
		// Check if a background refresh was triggered recently
		if was_background_refresh_triggered_recently() {
			return // Skip this refresh as one was triggered recently
		}

		// Try to read token
		token := read_token()
		if token.Error != "" {
			return // Can't authenticate without user interaction
		}

		spawn_background_refresh()
	}
}

// Function to force refresh the local cache
func refresh_local_cache() {
	// Try to read token, and initiate authentication mechanism if it fails
	token := read_token()
	if token.Error != "" {
		init_auth()
		return
	}

	// Check if token has expired
	time_location, _ := time.LoadLocation("UTC")
	time_format := "2006-01-02 15:04:05"
	token_time, err := time.Parse(time_format, token.CreationTime)
	if err != nil || time.Now().In(time_location).Sub(token_time).Milliseconds() > int64(token.Expires) {
		init_auth()
		return
	}

	// Check token lifetime and refresh if needed
	check_token_lifetime(token)

	// Force refresh all caches
	// Note: get_all_bookmarks will also refresh tags and collections when called with "fetch"
	get_all_bookmarks(token, "fetch")

	// Show a message to confirm the refresh
	wf.NewItem("Local caches have been refreshed").
		Subtitle("The caches now contain the latest bookmarks, tags, and collections from Raindrop.io").
		Valid(false)
}
