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
  $collection_info = explode ("⏦" , file_get_contents("current_collection.tmp"));
  $collection_search = true;
  $collection_search_name = $collection_info[1];
  $collection_search_id = (int)$collection_info[2];
  $collection_search_icon = $collection_info[3];
}
$workflow = new Workflow;

if ($collection_search) {
  $workflow->result()
    ->arg("⬅︎")
    ->title("Bookmarks in " . $collection_search_name)
    ->subtitle("⬅︎ Go back and search for all bookmarks")
    ->icon($collection_search_icon);
}

// Check if the token file exists and otherwise send the user over to the authentication
if (!file_exists("token.json")) {
  init_auth ($workflow);
}

// Read token and related data from file
$token = json_decode(file_get_contents("token.json"), true);

if ($query != "") {
// Query Raindrop.io
  $raindrop_results = search($query, $token["access_token"], $collection_search_id);

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
      $raindrop_results = search($query, $new_token["access_token"], $collection_search_id);
    }
  }

  // Search for collections that matches the search query, but only if we are not already doing a search in a collection
  if (!$collection_search) {
    // Get collection list from cache
    $raindrop_collections = array_reverse(collections($token["access_token"], false, "trust")["items"]);
    $raindrop_collections_sublevel = array_reverse(collections($token["access_token"], true, "trust")["items"]);

    // Render collections
    render_collections($raindrop_collections, $raindrop_collections_sublevel, $workflow, "paths", "searching");

    // Filter collections by search query
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

  // If we are inside a collection
  if ($collection_search) {
    $raindrop_results = search("", $token["access_token"], $collection_search_id);
  }
  // If we are are in standard search mode
  else {
    // Cache the collection list to make searching faster when the user types a search query
    array_reverse(collections($token["access_token"], false)["items"], "fetch");
    array_reverse(collections($token["access_token"], true)["items"], "fetch");

    // Default results if nothing is searched for. Just go to Raindrop.io itself
    $workflow->result()
      ->arg("https://app.raindrop.io/")
      ->title("Search your Raindrop.io bookmarks")
      ->subtitle("Or press enter to open Raindrop.io");
  }
}

if ($query != "" || $collection_search) {
  // Prepare results for being viewed in Alfred
  foreach ($raindrop_results["items"] as $result) {
    $workflow->result()
      ->uid("raindrop.io." . $result["_id"])
      ->arg($result["link"])
      ->title($result["title"])
      ->subtitle($result["excerpt"] != "" ? $result["excerpt"] : $result["link"])
      ->copy($result["link"])
      ->mod('cmd', $result["link"], $result["link"])
      ->mod('alt', "Press enter to copy this link to clipboard", "copy:::" . $result["link"]);
  }
}

// Output to Alfred
echo $workflow->output();