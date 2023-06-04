/*
	Functions for searching and browsing Raindrop.io bookmarks in Alfred

	By Andreas Westerlind in 2021
*/

package main

import (
	"encoding/json"
	"strconv"
	"strings"

	aw "github.com/deanishe/awgo"
)

func search(variant string, query string, collection_json string, tag string, from string, descr_in_list bool, favs_first bool) {
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

	// Try to read token, and initiate authentication mekanism if it fails
	token := read_token()
	if token.Error != "" {
		init_auth()
		return
	}

	var raindrop_results []interface{}

	var err error

	if query != "" {
		// Query Raindrop.io
		raindrop_results, err = search_request(query, token, collection_search_id, tag)
		if err != nil {
			new_token := refresh_token(token)
			if new_token.Error != "" {
				// Refreshing the token failed, so all we can do now is let the user authenticate again
				// We will also remove the old token, so that the Workflow knows that an authentication
				// is needed next time it's initiated
				wf.Keychain.Delete("raindrop_token")
				init_auth()
			} else {
				// Try to query Raindrop again, and assume it will work now as we just got a fresh new token to authenticate with
				token = new_token
				raindrop_results, _ = search_request(query, token, collection_search_id, tag)
			}
		}

		// Get collection list from cache
		raindrop_collections := reverse_interface_array(get_collections(token, false, "trust"))
		raindrop_collections_sublevel := reverse_interface_array(get_collections(token, true, "trust"))

		// Search for collections and tags that matches the search query, but only if we are not already doing a search in a collection or a tag
		if !collection_search && !tag_search {
			// Render collections
			var current_object []string
			render_collections(raindrop_collections, raindrop_collections_sublevel, "paths", "searching", 0, current_object, -1, "", "")

			// Get tag list from cache
			raindrop_tags := get_tags(token, "check")

			// Render tags
			for _, item_interface := range raindrop_tags {
				item := item_interface.(map[string]interface{})
				alfred_item := wf.NewItem(item["_id"].(string)).
					Var("current_tag", item["_id"].(string)).
					Var("goto", "tag").
					Valid(true).
					Icon(&aw.Icon{Value: "tag.png", Type: ""})
				alfred_item.Alt().
					Var("current_tag", item["_id"].(string)).
					Var("goto", "tag").
					Subtitle("")
			}

			// Filter collections and tags by search query
			wf.Filter(strings.ToLower(query))
		}
	} else {
		// We got no search query

		check_token_lifetime(token)

		// If we are searching for bookmarks inside a collection or with a specific tag
		if collection_search || tag_search {
			raindrop_results, _ = search_request("", token, collection_search_id, tag)
		} else {
			// If we are are in standard search mode

			// Cache collection and tag lists to make searching faster when the user types a search query
			get_collections(token, false, "check")
			get_collections(token, true, "check")

			// Default results if nothing is searched for. Just go to Raindrop.io itself
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
				Var("goto", "browse").
				Subtitle("").
				Valid(true).
				Icon(&aw.Icon{Value: "folder.png", Type: ""})
			alfred_item2.Alt().
				Var("goto", "browse").
				Subtitle("")
		}
	}

	if query != "" || collection_search || tag_search {
		// Get collection list from cache (REVERSING OF THE ARRAYS MIGHT NEED TO BE DONE HERE)
		raindrop_collections := reverse_interface_array(get_collections(token, false, "check"))
		raindrop_collections_sublevel := reverse_interface_array(get_collections(token, true, "check"))

		var current_object []string
		collection_names := collection_paths(raindrop_collections, raindrop_collections_sublevel, make(map[int]string), 0, current_object, -1)

		var render_favourites string = "all"

		// Prepare favourites for being viewed in Alfred (if favourites_first is enabled)
		if favs_first {
			render_results(raindrop_results, "only", collection_names, descr_in_list)
			render_favourites = "none"
		}

		// Prepare the rest of the results (or all results if favourites_first is disabled) for being viewed in Alfred
		render_results(raindrop_results, render_favourites, collection_names, descr_in_list)
	}
}

func browse(query string, full_collection_paths bool) {
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
		Var("goto", "collection").
		Subtitle("").
		Valid(true).
		Icon(&aw.Icon{Value: "folder.png", Type: ""})
	alfred_item2.Alt().
		Var("collection_info", "{\"icon\":\"folder.png\",\"id\":\"-1\",\"name\":\"Unsorted\"}").
		Var("goto", "collection").
		Subtitle("")

	render_style := "tree"
	if full_collection_paths {
		render_style = "paths"
	}

	var raindrop_collections []interface{}
	var raindrop_collections_sublevel []interface{}

	// Get collection list
	raindrop_collections = reverse_interface_array(get_collections(token, false, "check"))
	raindrop_collections_sublevel = reverse_interface_array(get_collections(token, true, "check"))

	// Render collections
	var current_object []string
	render_collections(raindrop_collections, raindrop_collections_sublevel, render_style, "searching", 0, current_object, -1, "", "")

	if query != "" {
		wf.Filter(query)
	}
}
