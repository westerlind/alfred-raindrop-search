/*
	Common functions for different tasks needed to handle Raindrop.io integration in Alfred

	By Andreas Westerlind in 2021
*/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
	"golang.org/x/net/html"
)

type RaindropToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	CreationTime string `json:"creation_time"`
	Expires      int    `json:"expires"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error,omitempty"`
}

func init_auth() {
	// Start mini web server that handles recieving the authentication code from Raindrop.io. It will run for a maximum of 20 minutes.
	cmd := exec.Command("./raindrop_alfred", "authserver", "&")
	cmd.Start()
	// Output info and authentication link to Alfred
	wf.NewItem("You are not authenticated with Raindrop.io").
		Arg("https://raindrop.io/oauth/authorize?redirect_uri=" + url.QueryEscape("http://127.0.0.1:11038") + "&client_id=5e46fab9b2fbaee7314687d8").
		Subtitle("Press enter to authenticate now").
		Valid(true)
}

func read_token() RaindropToken {
	token := RaindropToken{}
	keychain_token, err := wf.Keychain.Get("raindrop_token")
	if err == nil {
		json.Unmarshal([]byte(keychain_token), &token)
	} else {
		token.Error = "Failed to read token from Keychain"
	}
	return token
}

// Function for refreshing the authentication token
func refresh_token(token RaindropToken) RaindropToken {
	// Prepare POST variables
	post_variables := url.Values{}
	post_variables.Add("client_id", client_id)
	post_variables.Add("client_secret", client_secret)
	post_variables.Add("refresh_token", token.RefreshToken)
	post_variables.Add("grant_type", "refresh_token")

	resp, err := http.PostForm("https://raindrop.io/oauth/access_token", post_variables)
	if err != nil {
		token.Error = "Error making request"
		return token
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		token.Error = "Error reading response"
		return token
	}

	var new_token RaindropToken
	json.Unmarshal([]byte(body), &new_token)

	if new_token.Error != "" {
		return new_token
	}

	location, _ := time.LoadLocation("UTC")
	new_token.CreationTime = time.Now().In(location).Format("2006-01-02 15:04:05")

	token_json, _ := json.Marshal(new_token)

	// Save to Keychain
	if err := wf.Keychain.Set("raindrop_token", string(token_json)); err != nil {
		new_token.Error = "Failed to save token to Keychain"
		return new_token
	}

	return new_token
}

func search_request(query string, token RaindropToken, collection int, tag string) ([]interface{}, error) {
	// Prepare for searching by tag, if a tag is provided
	if tag != "" {
		tag = "{\"key\":\"tag\",\"val\":\"" + url.QueryEscape(tag) + "\"},"
	}

	// Sorting behaviour
	var sorting string = "score"
	if query == "" {
		sorting = "-created"
	}

	// Query Raindrop.io
	var result map[string]interface{}
	var err error
	client := &http.Client{}
	request, err := http.NewRequest("GET", "https://api.raindrop.io/rest/v1/raindrops/"+fmt.Sprint(collection)+"/?search=["+tag+"{\"key\":\"word\",\"val\":\""+url.QueryEscape(query)+"\"}]&sort=\""+sorting+"\"", nil)
	if err != nil {
		var nothing []interface{}
		return nothing, err
	}
	request.Header.Set("User-Agent", "Alfred (Macintosh; Mac OS X)")
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	response, err := client.Do(request)
	if err != nil {
		var nothing []interface{}
		return nothing, err
	}
	response_body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		var nothing []interface{}
		return nothing, err
	}

	json.Unmarshal(response_body, &result)

	result_items := result["items"].([]interface{})
	if result_items == nil {
		var nothing []interface{}
		err = errors.New("Could not get results")
		return nothing, err
	}

	return result_items, err
}

// Function for rendering Raindrop.io query results
func render_results(raindrop_results []interface{}, include_favourites string, collection_names map[int]string, descr_in_list bool) {
	for _, item_interface := range raindrop_results {
		item := item_interface.(map[string]interface{})

		var is_fav bool = false
		if _, ok := item["important"]; ok {
			if item["important"].(bool) == true {
				is_fav = true
			}
		}

		if include_favourites == "all" || (include_favourites == "none" && !is_fav) || (include_favourites == "only" && is_fav) {
			tag_list := ""
			tag_array := item["tags"].([]interface{})
			for _, current_tag := range tag_array {
				tag_list += "#" + current_tag.(string) + " "
			}
			if tag_list != "" {
				tag_list += " •  "
			}

			var fav_symbol string = ""
			if is_fav {
				fav_symbol = "♥︎ "
			}

			excerpt := item["excerpt"].(string)
			if excerpt == "" {
				excerpt = item["link"].(string)
			}

			// Prepare to display collection name
			var collection_name string
			if item["collection"] != nil {
				item_collection := item["collection"].(map[string]interface{})
				collection_name = collection_names[int(item_collection["$id"].(float64))]
			}
			if collection_name != "" {
				collection_name += " •  "
			}

			// collection_names missing so far
			subtitle_general := fav_symbol + collection_name + tag_list + get_hostname(item["link"].(string))
			subtitle_description := fav_symbol + excerpt
			var subtitle_main string
			var subtitle_alt string
			if descr_in_list {
				subtitle_main = subtitle_description
				subtitle_alt = subtitle_general
			} else {
				subtitle_main = subtitle_general
				subtitle_alt = subtitle_description
			}

			alfred_item := wf.NewItem(item["title"].(string)).
				Arg(string(item["link"].(string))).
				Var("goto", "open").
				Copytext(item["link"].(string)).
				Subtitle(subtitle_main).
				Valid(true)
			alfred_item.Cmd().
				Arg(item["link"].(string)).
				Var("goto", "open").
				Subtitle(item["link"].(string))
			alfred_item.Ctrl().
				Arg(string(item["link"].(string))).
				Var("goto", "open").
				Subtitle(subtitle_alt)
			alfred_item.Alt().
				//Arg("copy:::" + string(item["link"].(string))).
				Arg(string(item["link"].(string))).
				Var("goto", "copy").
				Subtitle("Press enter to copy this link to clipboard")
			alfred_item.Shift().
				Arg("https://api.raindrop.io/v1/raindrop/" + fmt.Sprint(int(item["_id"].(float64))) + "/cache").
				Subtitle("Press enter to open permantent copy")
		}
	}
}

// Function for getting Raindrop.io collections
func get_collections(token RaindropToken, sublevel bool, caching string) []interface{} {
	// If caching == "check": Redownload collection list only if cache is older than 1 minute, to make searching faster while still not having to wait for new collections to appear
	// If caching == "trust": Trust the collection list cache to be good enough and use what is cached without checking its age (only download if no chache exists yet)
	// If caching == "fetch": Always redownload collection list without checking the age of the cache

	var collections []interface{}
	var cache_base map[string]interface{}

	var cache_filename string = wf.CacheDir() + "/collections.json"
	if sublevel {
		cache_filename = wf.CacheDir() + "/collections_sublevel.json"
	}

	// Check if cache file exists
	if cache_file_stat, err := os.Stat(cache_filename); err == nil {
		// Ceck modification time of the cache file. Use cache if it is less than 1 minute old, and caching is set "check", or if caching is set to "trust"
		if caching == "trust" || (time.Now().Sub(cache_file_stat.ModTime()).Seconds() < 60 && caching == "check") {
			// Read stored cached collections
			cache_file, _ := os.ReadFile(cache_filename)
			json.Unmarshal(cache_file, &cache_base)
			if cache_base["items"] != nil && cache_base["items"].([]interface{}) != nil {
				collections = cache_base["items"].([]interface{})
				return collections
			}
		}
	}

	// Query Raindrop.io
	request_url := "https://api.raindrop.io/rest/v1/collections"
	if sublevel {
		request_url = "https://api.raindrop.io/rest/v1/collections/childrens"
	}
	client := &http.Client{}
	request, err := http.NewRequest("GET", request_url, nil)
	if err != nil {
		return collections
	}
	request.Header.Set("User-Agent", "Alfred (Macintosh; Mac OS X)")
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	response, err := client.Do(request)
	if err != nil {
		return collections
	}
	response_body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return collections
	}
	json.Unmarshal(response_body, &cache_base)

	// Write to file
	os.WriteFile(cache_filename, response_body, 0666)

	// Return collections
	if cache_base["items"] != nil && cache_base["items"].([]interface{}) != nil {
		collections = cache_base["items"].([]interface{})
	}
	return collections
}

// Returns only the hostname minus www from a given URL
func get_hostname(url_string string) string {
	url_object, _ := url.Parse(url_string)
	re := regexp.MustCompile("^www\\.")
	return re.ReplaceAllString(url_object.Host, "")
}

func collection_paths(raindrop_collections []interface{}, raindrop_collections_sublevel []interface{}, path_list map[int]string, parent_id int, current_object []string, current_level int) map[int]string {
	var collection_array []interface{}
	if parent_id == 0 {
		collection_array = raindrop_collections
	} else {
		collection_array = raindrop_collections_sublevel
	}

	for _, item_interface := range collection_array {
		item := item_interface.(map[string]interface{})
		var item_parent map[string]interface{}
		if item["parent"] != nil {
			item_parent = item["parent"].(map[string]interface{})
		}
		if parent_id == 0 || int(item_parent["$id"].(float64)) == parent_id {
			current_level++
			current_object = append(current_object, item["title"].(string))
			path_list[int(item["_id"].(float64))] = strings.Join(current_object, "/")

			path_list = collection_paths(raindrop_collections, raindrop_collections_sublevel, path_list, int(item["_id"].(float64)), current_object, current_level)

			// Remove last value from current_object
			if len(current_object) > 0 {
				current_object = current_object[:len(current_object)-1]
			}
			current_level--
		}
	}
	return path_list
}

// Function for rendering Raindrop.io collections in Alfred
func render_collections(raindrop_collections []interface{}, raindrop_collections_sublevel []interface{}, render_style string, purpose string, parent_id int, current_object []string, current_level int, bookmark_title string, bookmark_url string) {
	var collection_array []interface{}
	if parent_id == 0 {
		collection_array = raindrop_collections
	} else {
		collection_array = raindrop_collections_sublevel
	}

	for _, item_interface := range collection_array {
		item := item_interface.(map[string]interface{})
		var item_parent map[string]interface{}
		if item["parent"] != nil {
			item_parent = item["parent"].(map[string]interface{})
		}
		if parent_id == 0 || int(item_parent["$id"].(float64)) == parent_id {
			current_level++
			current_object = append(current_object, item["title"].(string))
			indentation := ""
			sub_indentation := ""
			if render_style == "tree" {
				if current_level > 0 {
					sub_indentation += "\t"
				}
				for i := 1; i < current_level; i++ {
					indentation += "\t"
					sub_indentation += "\t"
				}
				if current_level > 0 {
					indentation += "   ↳ "
					sub_indentation += "   "
				}
			}

			var icon_file_name = "folder.png"
			if len(item["cover"].([]interface{})) > 0 {
				icon_url_array := strings.Split(item["cover"].([]interface{})[0].(string), "/")
				if icon_url_array[len(icon_url_array)-1] != "" {
					icon_file_name = wf.CacheDir() + "/icon_cache/" + icon_url_array[len(icon_url_array)-1]
				}
			}

			// Make sure icon_cache folder exists
			os.MkdirAll(wf.CacheDir()+"/icon_cache", os.ModePerm)

			// Redownload the collection icon if the cached version is older than 60 days
			if strings.HasPrefix(icon_file_name, wf.CacheDir()+"/icon_cache/") {
				// Download/Redownload image if it doesnt exist or is more than 60 days old
				file_stat, err := os.Stat(icon_file_name)
				if err != nil || time.Now().Sub(file_stat.ModTime()).Hours() > 1440 {
					if !os.IsNotExist(err) {
						os.Remove(icon_file_name)
					}
					file, _ := os.Create(icon_file_name)
					defer file.Close()
					resp, _ := http.Get(item["cover"].([]interface{})[0].(string))
					defer resp.Body.Close()
					io.Copy(file, resp.Body)
				}
			}

			collection_title := item["title"].(string)
			if render_style == "paths" {
				collection_title = strings.Join(current_object, "/")
			}

			tree_arg_section := ""
			if render_style == "tree" {
				tree_arg_section = strings.ToLower(sub_collection_names(raindrop_collections_sublevel, int(item["_id"].(float64))))
			}

			collection_info := make(map[string]string)
			if purpose == "adding" {
				collection_info["collection"] = fmt.Sprint(int(item["_id"].(float64)))
				collection_info["title"] = bookmark_title
				collection_info["url"] = bookmark_url
			} else if purpose == "searching" {
				collection_info["id"] = fmt.Sprint(int(item["_id"].(float64)))
				collection_info["name"] = strings.Join(current_object, "/")
				collection_info["icon"] = icon_file_name
			}
			collection_json, _ := json.Marshal(collection_info)

			if purpose == "adding" {
				alfred_item := wf.NewItem(indentation+collection_title).
					Arg(fmt.Sprint(strings.ToLower(strings.Join(current_object, " "))+" "+tree_arg_section)).
					Var("bookmark_info", string(collection_json)).
					Valid(true).
					Icon(&aw.Icon{icon_file_name, ""})
				alfred_item.Cmd().
					Var("bookmark_info", string(collection_json)).
					Var("goto", "save now").
					Subtitle(sub_indentation + "Save now, without setting custom title or adding tags")
				alfred_item.Alt().
					Var("bookmark_info", string(collection_json)).
					Subtitle("")
			} else if purpose == "searching" {
				alfred_item := wf.NewItem(indentation+collection_title).
					Arg(strings.ToLower(strings.Join(current_object, " "))+" "+tree_arg_section).
					Var("collection_info", string(collection_json)).
					Var("goto", "collection").
					Valid(true).
					Icon(&aw.Icon{icon_file_name, ""})
				alfred_item.Alt().
					Arg(strings.ToLower(strings.Join(current_object, " "))+" "+tree_arg_section).
					Var("collection_info", string(collection_json)).
					Var("goto", "collection").
					Subtitle("")
			}

			render_collections(raindrop_collections, raindrop_collections_sublevel, render_style, purpose, int(item["_id"].(float64)), current_object, current_level, bookmark_title, bookmark_url)

			// Remove last value from current_object
			if len(current_object) > 0 {
				current_object = current_object[:len(current_object)-1]
			}
			current_level--
		}
	}
}

// Function for getting the names of all sub collections in a string
func sub_collection_names(raindrop_collections_sublevel []interface{}, parent_id int) string {
	names := ""
	for _, item_interface := range raindrop_collections_sublevel {
		item := item_interface.(map[string]interface{})
		var item_parent map[string]interface{}
		if item["parent"] != nil {
			item_parent = item["parent"].(map[string]interface{})
		}
		if int(item_parent["$id"].(float64)) == parent_id {
			names += item["title"].(string) + " " + sub_collection_names(raindrop_collections_sublevel, int(item["_id"].(float64)))
		}
	}
	return names
}

// Function for getting Raindrop.io tags
func get_tags(token RaindropToken, caching string) []interface{} {
	// If caching == "check": Redownload tag list only if cache is older than 1 minute, to make searching faster while still not having to wait for new tags to appear
	// If caching == "trust": Trust the tag list cache to be good enough and use what is cached without checking its age (only download if no chache exists yet)
	// If caching == "fetch": Always redownload tag list without checking the age of the cache

	var tags []interface{}
	var cache_base map[string]interface{}

	// Check if cache file exists
	if cache_file_stat, err := os.Stat("tags.json"); err == nil {
		// Ceck modification time of the cache file. Use cache if it is less than 1 minute old, and caching is set "check", or if caching is set to "trust"
		if caching == "trust" || (time.Now().Sub(cache_file_stat.ModTime()).Seconds() < 60 && caching == "check") {
			// Read stored cached collections
			cache_file, _ := os.ReadFile(wf.CacheDir() + "/tags.json")
			json.Unmarshal(cache_file, &cache_base)
			if cache_base["items"] != nil && cache_base["items"].([]interface{}) != nil {
				tags = cache_base["items"].([]interface{})
				return tags
			}
		}
	}

	// Query Raindrop.io
	request_url := "https://api.raindrop.io/rest/v1/tags/0"

	client := &http.Client{}
	request, err := http.NewRequest("GET", request_url, nil)
	if err != nil {
		return tags
	}
	request.Header.Set("User-Agent", "Alfred (Macintosh; Mac OS X)")
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	response, err := client.Do(request)
	if err != nil {
		return tags
	}
	response_body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return tags
	}
	json.Unmarshal(response_body, &cache_base)

	// Write to file
	os.WriteFile(wf.CacheDir()+"/tags.json", response_body, 0666)

	// Return tags
	if cache_base["items"] != nil && cache_base["items"].([]interface{}) != nil {
		tags = cache_base["items"].([]interface{})
	}
	return tags
}

func reverse_interface_array(array []interface{}) []interface{} {
	new_array := make([]interface{}, len(array))
	for count_up, count_down := 0, len(array)-1; count_up < len(array); count_up, count_down = count_up+1, count_down-1 {
		new_array[count_up] = array[count_down]
	}
	return new_array
}

// Check if Token has gone through more than half of it's lifetime, and in that case, refresh it
func check_token_lifetime(token RaindropToken) {
	time_location, _ := time.LoadLocation("UTC")
	time_format := "2006-01-02 15:04:05"
	token_time, _ := time.Parse(time_format, token.CreationTime)
	time_difference := time.Now().In(time_location).Sub(token_time).Milliseconds()
	if float64(time_difference) > float64(token.Expires)*0.5 {
		refresh_token(token)
	}
}

// Function for getting HTML meta description from a given URL
func get_meta_description(url_string string) string {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url_string, nil)
	if err != nil {
		return ""
	}
	request.Header.Set("User-Agent", "Alfred (Macintosh; Mac OS X)")
	response, err := client.Do(request)
	if err != nil {
		return ""
	}
	z := html.NewTokenizer(response.Body)
	defer response.Body.Close()
	for {
		z.Next()
		current_token := z.Token()
		//fmt.Println(current_token.Attr)
		if current_token.Data == "meta" {
			var found bool
			var description string
			for _, current_attribute := range current_token.Attr {
				if current_attribute.Key == "name" && current_attribute.Val == "description" {
					found = true
				}
				if current_attribute.Key == "content" {
					description = current_attribute.Val
				}
			}
			if found && description != "" {
				return description
			}
		}
		if z.Err() != nil {
			return ""
		}
	}
}

// Function for getting HTML title from a given URL
func get_title(url_string string) string {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url_string, nil)
	if err != nil {
		return ""
	}
	request.Header.Set("User-Agent", "Alfred (Macintosh; Mac OS X)")
	response, err := client.Do(request)
	if err != nil {
		return ""
	}
	z := html.NewTokenizer(response.Body)
	defer response.Body.Close()
	current_tag := ""
	for {
		tt := z.Next()
		current_token := z.Token()
		switch tt {
		case html.ErrorToken:
			return ""
		case html.TextToken:
			if current_tag == "title" {
				return current_token.String()
			}
		case html.StartTagToken:
			current_tag = current_token.Data
		case html.EndTagToken:
			if current_tag == "title" {
				return ""
			}
		}
	}
}

// Function for handling Firefox related error messages
func firefox_error(message string) {
	if message == "Workflow with Id 'net.deanishe.alfred.firefox-assistant' is disabled" {
		alfred_item := wf.NewItem("Firefox Assistant is disabled").
			Subtitle("Enable it in Alfred's preferences to be able to add bookmarks from Firefox").
			Valid(false).
			Icon(&aw.Icon{"firefox.png", ""})
		alfred_item.Alt().
			Subtitle("Enable it in Alfred's preferences to be able to add bookmarks from Firefox")
	} else if message == "Cannot find workflow with Id 'net.deanishe.alfred.firefox-assistant'" {
		alfred_item := wf.NewItem("Firefox Assistant is needed to add from Firefox").
			Arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md").
			Subtitle("Press enter for instructions to install and configure Firefox Assistant").
			Valid(true).
			Icon(&aw.Icon{"firefox.png", ""})
		alfred_item.Alt().
			Arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md").
			Subtitle("Press enter for instructions to install and configure Firefox Assistant")
	} else if message == "Cannot Connect to Extension" {
		alfred_item := wf.NewItem("Cannot Connect to Firefox Extension").
			Arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md").
			Subtitle("Press enter for instructions to install and configure Firefox Assistant").
			Valid(true).
			Icon(&aw.Icon{"firefox.png", ""})
		alfred_item.Alt().
			Arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md").
			Subtitle("Press enter for instructions to install and configure Firefox Assistant")
	} else if message == "Failed to read information from Firefox" {
		alfred_item := wf.NewItem("Failed to read information from Firefox").
			Subtitle("It will probably work if you try again").
			Icon(&aw.Icon{"firefox.png", ""})
		alfred_item.Alt().
			Arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md").
			Subtitle("Press enter for instructions to install and configure Firefox Assistant")
	} else {
		alfred_item := wf.NewItem("Something went wrong while connecting to Firefox").
			Arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md").
			Subtitle("Press enter for instructions to install and configure Firefox Assistant").
			Valid(true).
			Icon(&aw.Icon{"firefox.png", ""})
		alfred_item.Alt().
			Arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md").
			Subtitle("Press enter for instructions to install and configure Firefox Assistant")
	}

	wf.Var("ff_error", "true")
}
