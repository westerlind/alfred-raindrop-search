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

$render_style = "tree";
if ($argv[2] == "true") {
  $render_style = "paths";
}

$workflow = new Workflow;

$workflow->result()
  ->arg("⬅︎")
  ->title("Raindrop.io Bookmark Collections")
  ->subtitle("⬅︎ Go back to search all bookmarks")
  ->icon("icon.png")
  ->mod("alt", "⬅︎ Go back to search all bookmarks", "⬅︎");

// Read token and related data from file
$token = json_decode(file_get_contents("token.json"), true);

if ($query != "") {
  // Get collection list from cache
  $raindrop_collections = array_reverse(collections($token["access_token"], false, "trust")["items"]);
  $raindrop_collections_sublevel = array_reverse(collections($token["access_token"], true, "trust")["items"]);
}
else {
  // Get collection list from Raindrop.io
  $raindrop_collections = array_reverse(collections($token["access_token"], false, "check")["items"]);
  $raindrop_collections_sublevel = array_reverse(collections($token["access_token"], true, "check")["items"]);
}

// Render collections
render_collections($raindrop_collections, $raindrop_collections_sublevel, $workflow, $render_style, "searching");

if ($query != "") {
  // Filter collections by search query
  $workflow->filterResults(mb_strtolower($query), 'arg');
}

// Output to Alfred
echo $workflow->output();
