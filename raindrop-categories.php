<?php
// Script for getting and rendering Raindrop.io collections for Alfred,
// and preparing for adding a new bookmark
//
// By Andreas Westerlind in 2020
//

require './WorkflowsPHPHelper/src/Workflow.php';
require './WorkflowsPHPHelper/src/Result.php';
require 'raindrop-common.php';
use Alfred\Workflows\Workflow;

$query = $argv[2];
if (substr($argv[1], 0, 16) === "{\"alfredworkflow") {
  $firefox_tab = json_decode($argv[1], true);
  $browserUrl = $firefox_tab["alfredworkflow"]["variables"]["FF_URL"];
  $browserTitle = $firefox_tab["alfredworkflow"]["variables"]["FF_TITLE"];
}
else {
  $browserUrl = $argv[1];
  $browserTitle = $argv[3];
}
$workflow = new Workflow;

// Check if the token file exists and otherwise send the user over to the authentication
if (!file_exists("token.json")) {
  init_auth($workflow);
}

// Read token and related data from file
$token = json_decode(file_get_contents("token.json"), true);

if ($query == "") {
  // Check if Token has gone through more than half of it's lifetime, and in that case, refresh it
  $current_time = new DateTime("now", new DateTimeZone('UTC'));
  $token_time = date_create_from_format("Y-m-d H:i:s", $token["creation_time"], new DateTimeZone('UTC'));
  $time_difference = $token_time->diff($current_time);
  if ((int)$token["expires"] - date_interval_to_milliseconds($time_difference) < (int)$token["expires"] * 0.5) {
    refresh_token($token["refresh_token"]);
  }
}

if ($browserUrl === "No browser active") {
  // Result we didn't get any URL to save, probably because no browser is the frontmost app
  $workflow->result()
    ->valid(false)
    ->title("There is nothing here to add to Raindrop.io")
    ->subtitle("Go to the browser you want to add a bookmark from and try again");
  echo $workflow->output();
  die();
}

// Put alternative to add the new bookmark to Unsorted above the category list
$workflow->result()
  ->arg("")
  ->mod('cmd', $sub_indentation . "Open Raindrop.io to change details after saving", " :§:open_raindrop:§: ")
  ->title("Add Raindrop.io Bookmark to Unsorted")
  ->subtitle("Or select a collection below");

// Make sure that the icon_cache directory exists
if (!file_exists('icon_cache')) {
  mkdir('icon_cache', 0777, true);
}

// Get categories
$raindrop_categories = array_reverse(categories($token["access_token"], false)["items"]);
$raindrop_categories_sublevel = array_reverse(categories($token["access_token"], true)["items"]);

// Render categories
render_categories($raindrop_categories, $raindrop_categories_sublevel, $workflow);

// Add Alfred variable for the URL we want to add to Raindrop
$workflow->variable('url', $browserUrl);
$workflow->variable('title', $browserTitle);

// Output to Alfred
if ($query == "") {
  //echo $workflow->sortResults('asc', 'uid')->output();
  echo $workflow->output();
} else {
  //echo $workflow->filterResults(mb_strtolower($query), 'uid')->sortResults('asc', 'uid')->output();
  echo $workflow->filterResults(mb_strtolower($query), 'arg')->output();
}

// ----------FUNCTIONS----------

// Function for rendering Raindrop.io categories in Alfred
function render_categories($raindrop_categories, $raindrop_categories_sublevel, $workflow, $parent_id = 0, $current_object = [], $current_level = -1)
{
  if ($parent_id == 0) {
    $category_array = $raindrop_categories;
  } else {
    $category_array = $raindrop_categories_sublevel;
  }
  foreach ($category_array as $result) {
    if ($parent_id == 0 || $result["parent"]["\$id"] === $parent_id) {
      $current_level++;
      $current_object[$current_level] = mb_strtolower($result["title"]);

      $indentation = "";
      $sub_indentation = "";
      if ($current_level > 0) {
        $indentation .= " ";
        $sub_indentation .= "          ";
      }
      for ($i = 1; $i < $current_level; $i++) {
        $indentation .= "      ";
        if ($current_level == 3) {
          $sub_indentation .= "        ";
        }
        else {
          $sub_indentation .= "         ";
        }
      }
      if ($current_level > 0) {
        $indentation .= " ↳ ";
      }

      $icon_url_array = explode("/", $result["cover"][0]);
      $icon_file_name = "icon_cache/" . $icon_url_array[key(array_slice($icon_url_array, -1, 1, true))];
      if (!file_exists($icon_file_name)) {
        $icon_content = file_get_contents($result["cover"][0]);
        file_put_contents($icon_file_name, $icon_content);
      }

      $workflow->result()
        ->arg($result["_id"] . " " . implode(" ", $current_object) . " " . mb_strtolower(sub_category_names($raindrop_categories_sublevel, $result["_id"])))
        ->mod('cmd', $sub_indentation."Open Raindrop.io to change details after saving", $result["_id"] . " :§:open_raindrop:§: " . mb_strtolower(sub_category_names($raindrop_categories_sublevel, $result["_id"])))
        ->icon($icon_file_name)
        ->title($indentation . $result["title"]);

      render_categories($raindrop_categories, $raindrop_categories_sublevel, $workflow, $result["_id"], $current_object, $current_level);

      unset($current_object[$current_level]);
      $current_level--;
    }
  }
}

// Function for getting the names of all sub categories in a string
function sub_category_names($raindrop_categories_sublevel, $parent_id) {
  $names = "";
  foreach ($raindrop_categories_sublevel as $result) {
    if ($result["parent"]["\$id"] === $parent_id) {
      $names .= $result["title"]." ".sub_category_names($raindrop_categories_sublevel, $result["_id"]);
    }
  }
  return $names;
}

// Function for getting Raindrop.io categories
function categories(string $token, bool $sublevel) {
  if (file_exists($sublevel ? "categories_sublevel.json" : "categories.json") && time() - filemtime($sublevel ? "categories_sublevel.json" : "categories.json") < 60) {
    // Read stored cached categories
    $raindrop_results = json_decode(file_get_contents($sublevel ? "categories_sublevel.json" : "categories.json"), true);
    if ($raindrop_results["result"] == 1) {
      return $raindrop_results;
    }
  }

  // Query Raindrop.io
  $curl = curl_init();
  if ($sublevel) {
    curl_setopt($curl, CURLOPT_URL, "https://api.raindrop.io/rest/v1/collections/childrens");
  } else {
    curl_setopt($curl, CURLOPT_URL, "https://api.raindrop.io/rest/v1/collections");
  }
  curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
  curl_setopt($curl, CURLOPT_USERAGENT, "Alfred (Macintosh; Mac OS X)");
  curl_setopt($curl, CURLOPT_HTTPHEADER, array('Authorization: Bearer ' . $token));
  $raindrop_results = json_decode(curl_exec($curl), true);
  curl_close($curl);
  file_put_contents($sublevel ? "categories_sublevel.json" : "categories.json", json_encode($raindrop_results));
  return $raindrop_results;
}
