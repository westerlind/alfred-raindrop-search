<?php
// Script for searching Raindrop.io bookmarks with Alfred,
// and handling parts of the OAuth authentication process
//
// By Andreas Westerlind in 2020
//

require './WorkflowsPHPHelper/src/Workflow.php';
require './WorkflowsPHPHelper/src/Result.php';
require 'raindrop-common.php';
use Alfred\Workflows\Workflow;

$query = $argv[1];

$collection_search = false;
$collection_search_id = 0;
if ($argv[2] == "collection") {
  $collection_info_store = file_get_contents("current_collection.tmp");
  $collection_info = explode ("⏦" , $collection_info_store);
  $collection_search = true;
  $collection_search_name = $collection_info[1];
  $collection_search_id = (int)$collection_info[2];
  $collection_search_icon = $collection_info[3];
}

$tag_search = false;
$tag = "";
if ($argv[2] == "tag") {
  $tag = file_get_contents("current_tag.tmp");
  $tag_search = true;
}

$prefer_description = false;
if ($argv[3] == "true") {
  $prefer_description = true;
}

$workflow = new Workflow;

if ($collection_search) {
  // We are browsing a collection, and came here from the collection browser
  if (mb_substr($collection_info_store, -18, 17) == "§collection_list§") {
    $workflow->result()
      ->arg("⬅︎⊟")
      ->title("Bookmarks in " . $collection_search_name)
      ->subtitle("⬅︎ Go back to collection browser")
      ->icon($collection_search_icon)
      ->mod("alt", "⬅︎ Go back to collection browser", "⬅︎⊟");
  }
  // We are browsing a collection, and came here from the main bookmark search
  else {
    $workflow->result()
      ->arg("⬅︎")
      ->title("Bookmarks in " . $collection_search_name)
      ->subtitle("⬅︎ Go back to search all bookmarks")
      ->icon($collection_search_icon)
      ->mod("alt", "⬅︎ Go back to search all bookmarks", "⬅︎");
  }
}

if ($tag_search) {
  // We are browsing bookmarks with a specific tag
  $workflow->result()
    ->arg("⬅︎")
    ->title("Bookmarks tagged with #" . $tag)
    ->subtitle("⬅︎ Go back to search all bookmarks")
    ->icon("tag.png")
    ->mod("alt", "⬅︎ Go back to search all bookmarks", "⬅︎");
}

// Check if the token file exists and otherwise send the user over to the authentication
if (!file_exists("token.json")) {
  init_auth ($workflow);
}

// Read token and related data from file
$token = json_decode(file_get_contents("token.json"), true);

if ($query != "") {
// Query Raindrop.io
  $raindrop_results = search($query, $token["access_token"], $collection_search_id, $tag);

  if (!isset($raindrop_results["items"]) && isset($raindrop_results["result"]) && !$raindrop_results["result"]) {
    // We got an error instead of the content we wanted, and that's probably because the token is outdated, so first try to refresh it
    $new_token = refresh_token($token["refresh_token"]);
    if(!$new_token) {
      // Refreshing the token failed, so all we can do now is let the user authenticate again
      // We will also remove the old token-file, so that the Workflow knows that an authentication 
      // is needed next time it's initiated
      unlink("token.json");
      init_auth($workflow);
    }
    else {
      // Try to query Raindrop again, and assume it will work now as we just got a fresh new token to authenticate with
      $raindrop_results = search($query, $new_token["access_token"], $collection_search_id, $tag);
    }
  }

  // Get collection list from cache
  $raindrop_collections = array_reverse(collections($token["access_token"], false, "trust")["items"]);
  $raindrop_collections_sublevel = array_reverse(collections($token["access_token"], true, "trust")["items"]);

  // Search for collections and tags that matches the search query, but only if we are not already doing a search in a collection or a tag
  if (!$collection_search && !$tag_search) {
    // Render collections
    render_collections($raindrop_collections, $raindrop_collections_sublevel, $workflow, "paths", "searching");

    // Get tag list from cache
    $raindrop_tags = tags($token["access_token"], "trust")["items"];

    // Render tags
    foreach ($raindrop_tags as $current_tag) {
      $workflow->result()
        ->arg("⌈" . $current_tag["_id"] . "⌈")
        ->title($current_tag["_id"])
        ->icon("tag.png")
        ->mod("alt", "", "⌈" . $current_tag["_id"] . "⌈");
    }

    // Filter collections and tags by search query
    $workflow->filterResults(mb_strtolower($query), 'arg');
  }
}
else {
  // We got no search query

  // Check if Token has gone through more than half of it's lifetime, and in that case, refresh it
  $current_time = new DateTime("now", new DateTimeZone('UTC'));
  $token_time = date_create_from_format("Y-m-d H:i:s", $token["creation_time"], new DateTimeZone('UTC'));
  $time_difference = $token_time->diff($current_time);
  if ((int)$token["expires"] - date_interval_to_milliseconds($time_difference) < (int)$token["expires"] * 0.5) {
    refresh_token($token["refresh_token"]);
  }

  // If we are searching for bookmarks inside a collection or with a specific tag
  if ($collection_search || $tag_search) {
    $raindrop_results = search("", $token["access_token"], $collection_search_id, $tag);

    // Get collection list from cache
    $raindrop_collections = array_reverse(collections($token["access_token"], false, "trust")["items"]);
    $raindrop_collections_sublevel = array_reverse(collections($token["access_token"], true, "trust")["items"]);
  }
  // If we are are in standard search mode
  else {
    // Cache collection and tag lists to make searching faster when the user types a search query
    $raindrop_collections = collections($token["access_token"], false, "check");
    $raindrop_collections_sublevel = collections($token["access_token"], true, "check");
    tags($token["access_token"], "check");

    // Default results if nothing is searched for. Just go to Raindrop.io itself
    $workflow->result()
      ->arg("https://app.raindrop.io/")
      ->title("Search your Raindrop.io bookmarks")
      ->subtitle("Or press enter to open Raindrop.io")
      ->mod("alt", "Or press enter to open Raindrop.io", "https://app.raindrop.io/");
    $workflow->result()
      ->arg("browse➡︎")
      ->title("Browse your Raindrop.io collections")
      ->icon("folder.png")
      ->mod("alt", "", "browse➡︎");
  }
}

if ($query != "" || $collection_search || $tag_search) {

  $collection_names = collection_paths($raindrop_collections, $raindrop_collections_sublevel);

  // Prepare results for being viewed in Alfred
  foreach ($raindrop_results["items"] as $result) {
    $tag_list = "";
    foreach ($result["tags"] as $current_tag) {
      $tag_list .= "#" . $current_tag . " ";
    }
    if ($tag_list != "") {
      $tag_list .= " •  ";
    }

    $workflow->result()
      ->arg($result["link"])
      ->title($result["title"])
      ->subtitle($prefer_description ? ($result["excerpt"] != "" ? $result["excerpt"] : $result["link"]) : $collection_names[$result["collection"]["\$id"]] . "  •  " . $tag_list . preg_replace('/^www\./', '', parse_url($result["link"])["host"]))
      ->copy($result["link"])
      ->mod('cmd', $result["link"], $result["link"])
      ->mod('ctrl', $prefer_description ? $collection_names[$result["collection"]["\$id"]] . "  •  " . $tag_list . preg_replace('/^www\./', '', parse_url($result["link"])["host"]) : ($result["excerpt"] != "" ? $result["excerpt"] : "No description"), $result["link"])
      ->mod('alt', "Press enter to copy this link to clipboard", "copy:::" . $result["link"]);
  }
}

// Output to Alfred
echo $workflow->output();