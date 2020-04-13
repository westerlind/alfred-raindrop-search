<?php
// Script for supporting the process of adding tags to a
// new bookmark that is added to Raindrop.io via Alfred
//
// By Andreas Westerlind in 2020
//

require './WorkflowsPHPHelper/src/Workflow.php';
require './WorkflowsPHPHelper/src/Result.php';
require 'raindrop-common.php';

use Alfred\Workflows\Workflow;

$workflow = new Workflow;

$query = $argv[1];
$tags = explode(",", $query);
foreach ($tags as $key => $current_tag) {
  $tags[$key] = trim(trim($current_tag), "#");
}

// Read token and related data from file
$token = json_decode(file_get_contents("token.json"), true);

$filtered_tags = [];

if ($query != "" && $tags[count($tags) - 1] != "") {
  // Get tag list from cache
  $raindrop_tags = tags($token["access_token"], "trust")["items"];

  foreach ($raindrop_tags as $current_tag) {
    if(strpos($current_tag["_id"], $tags[count($tags) - 1]) !== false) {
      $filtered_tags[] = $current_tag;
    }
  }
} else {
  // Get tag list from Raindrop.io and cache it
  tags($token["access_token"], "check");
}

$tag_list = "";
$valid_tag_count = 0;
foreach ($tags as $current_tag) {
  if ($current_tag != "") {
    $tag_list .= "#" . $current_tag . " ";
    $valid_tag_count++;
  }
}

$previous_tags = "";
$pos = strrpos($query, ',');
if ($pos !== false) {
  $previous_tags = substr($query, 0, $pos) . ", ";
}

$tag_info = "Save without tags ";
if($valid_tag_count == 1) {
  $tag_info = "Save with tag ";
}
if($valid_tag_count > 1) {
  $tag_info = "Save with tags ";
}

$workflow->result()
  ->arg($query)
  ->title($tag_info . $tag_list)
  ->subtitle("Separate multiple tags with comma: tag1, tag2, tag3")
  ->mod("alt", "Separate multiple tags with comma: tag1, tag2, tag3", $query);

if ($query != "") {
  foreach ($filtered_tags as $current_tag) {
    $workflow->result()
      ->arg("-⟲" . $previous_tags . $current_tag["_id"] . ", ")
      ->title($current_tag["_id"])
      ->mod("cmd", "Add this tag and save",  $previous_tags . $current_tag["_id"])
      ->icon("tag.png")
      ->mod("alt", "", "-⟲" . $previous_tags . $current_tag["_id"] . ", ");
  }
}

// Output to Alfred
echo $workflow->output();