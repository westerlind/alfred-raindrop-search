/*
	By Andreas Westerlind, 2021-2025
*/

package main

import (
	"flag"
	"os"

	aw "github.com/deanishe/awgo"
)

var wf *aw.Workflow

func init() {
	wf = aw.New()
}

func run() {
	// Handle command line arguments
	f := os.Args[1]
	var query string
	var variant string
	var bookmark_url string
	var bookmark_title string
	var firefox_json string
	var message string
	var title string
	var tags string
	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	flagSet.StringVar(&query, "query", "", "Search Query")
	flagSet.StringVar(&variant, "variant", "standard", "Variant of the main selected function")
	flagSet.StringVar(&bookmark_url, "bookmark_url", "", "URL of bookmark that should be added")
	flagSet.StringVar(&bookmark_title, "bookmark_title", "", "Title of bookmark that should be added")
	flagSet.StringVar(&firefox_json, "firefox_json", "", "Firefox json")
	flagSet.StringVar(&message, "message", "", "Message, for example forwarded error message to handle")
	flagSet.StringVar(&title, "title", "", "Bookmark title")
	flagSet.StringVar(&tags, "tags", "", "Comma separated bookmark tags")
	flagSet.Parse(os.Args[2:])
	descr_in_list := false
	favs_first := true
	full_collection_paths := false
	if wf.Config.Get("description_in_boomark_listing", "0") == "1" {
		descr_in_list = true
	}
	if wf.Config.Get("favourites_first", "true") == "0" {
		favs_first = false
	}
	if wf.Config.Get("subcollections_as_full_paths", "false") == "1" {
		full_collection_paths = true
	}

	// Select function
	if f == "search" {
		search(variant, query, wf.Config.Get("collection_info", ""), tags, wf.Config.Get("from", ""), descr_in_list, favs_first)
	}
	if f == "local_search" {
		local_search_command(variant, query, wf.Config.Get("collection_info", ""), tags, wf.Config.Get("from", ""), descr_in_list, favs_first)
	}
	if f == "refresh_cache" {
		refresh_local_cache()
	}
	if f == "browse" {
		browse(query, full_collection_paths)
	}
	if f == "local_browse" {
		local_browse(query, full_collection_paths)
	}
	if f == "select_collection" {
		select_collection(query, bookmark_url, bookmark_title, firefox_json, full_collection_paths)
	}
	if f == "firefox_error" {
		firefox_error(message)
	}
	if f == "set_title" {
		set_title(title)
	}
	if f == "set_tags" {
		set_tags(tags)
	}

	wf.SendFeedback()
}

func main() {
	if os.Args[1] == "authserver" {
		// If the first argument is "authserver", start the authserver
		authserver()
	} else if os.Args[1] == "background_refresh" {
		// If the first argument is "background_refresh", refresh the cache in the background
		background_refresh_cache()
	} else if os.Args[1] == "save_bookmark" {
		// If the first argument is "save_bookmark", then go and save the bookmark
		var tags string
		flagSet := flag.NewFlagSet("", flag.ExitOnError)
		flagSet.StringVar(&tags, "tags", "", "Comma separated bookmark tags")
		flagSet.Parse(os.Args[2:])
		save_bookmark(tags)
	} else if os.Args[1] == "logout" {
		// If the first argument is "logout", remove the token from the Keychain
		logout()
	} else {
		// Else, run normally, meaning that we will run with the assumption that we will output json for rendering in Alfred
		wf.Run(run)
	}
}
