<?php
// Common functions for the different parts of this Raindrop.io Alfred Workflow
//
// By Andreas Westerlind in 2020
//

// Function for initiating authentication process
function init_auth($workflow)
{
  // Start Python web server for handling recieving the authentication code from Raindrop.io. It will run for a maximum of 20 minutes.
  shell_exec("screen -dmS raindrop_authserver python raindrop_authserver.py");
  // Output info and authentication link to Alfred
  $workflow->result()
    ->arg("https://raindrop.io/oauth/authorize?redirect_uri=" . urlencode("http://127.0.0.1:11038") . "&client_id=5e46fab9b2fbaee7314687d8")
    ->title("You are not authenticated with Raindrop.io")
    ->subtitle("Press enter to authenticate now");
  echo $workflow->output();
  die();
}

// Function for refreshing the authentication token
function refresh_token(string $refresh_token)
{
  // Get client_id and client_secret from client_code.json
  // Those values are specific to this Alfred Workflow.
  // Please get your own codes from https://app.raindrop.io/#/settings/apps/dev if you are going to use this script for something else!
  $client_code = json_decode(file_get_contents("client_code.json"), true);

  // Prepare POST variables
  $post_variables = array(
    "client_id" => $client_code["client_id"],
    "client_secret" => $client_code["client_secret"],
    "refresh_token" => $refresh_token,
    "grant_type" => "refresh_token"
  );

  // Query Raindrop.io for refreshed token
  $curl = curl_init();
  curl_setopt($curl, CURLOPT_URL, "https://raindrop.io/oauth/access_token");
  curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
  curl_setopt($curl, CURLOPT_USERAGENT, "Alfred (Macintosh; Mac OS X)");
  curl_setopt($curl, CURLOPT_POSTFIELDS, http_build_query($post_variables));
  $token = json_decode(curl_exec($curl), true);
  curl_close($curl);

  if ($token["error"]) {
    return false;
  } else {
    $creation_time = new DateTime("now", new DateTimeZone('UTC'));
    $token["creation_time"] = $creation_time->format("Y-m-d H:i:s");
    file_put_contents("token.json", json_encode($token));
    return $token;
  }
}

// Function for getting total amount of milliseconds from DateInterval variable
function date_interval_to_milliseconds(DateInterval $interval)
{
  $milliseconds = (int) $interval->format("%f") / 1000;
  $milliseconds += (int) $interval->format("%s") * 1000;
  $milliseconds += (int) $interval->format("%i") * 60 * 1000;
  $milliseconds += (int) $interval->format("%h") * 60 * 60 * 1000;
  $milliseconds += (int) $interval->format("%a") * 24 * 60 * 60 * 1000;
  return $milliseconds;
}

// Function for searching an array for a partial string. Returns true if found or false if not.
function partial_string_in_array($needle, $haystack)
{
  foreach ($haystack as $current) {
    if (strpos($current, $needle) !== false) {
      return true;
    }
  }
  return false;
}

// Function for searching Raindrop.io bookmarks
function search(string $query, string $token, int $collection = 0, string $tag = "")
{
  // Prepare for searching by tag, if a tag is provided
  if($tag != "") {
    $tag = "{\"key\":\"tag\",\"val\":\"" . urlencode($tag) . "\"},";
  }

  // Query Raindrop.io
  $curl = curl_init();
  curl_setopt($curl, CURLOPT_URL, "https://api.raindrop.io/rest/v1/raindrops/" . $collection . "/?search=[" . $tag . "{\"key\":\"word\",\"val\":\"" . urlencode($query) . "\"}]&sort=\"" . ($query == "" ? "-created" : "score")."\"");
  curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
  curl_setopt($curl, CURLOPT_USERAGENT, "Alfred (Macintosh; Mac OS X)");
  curl_setopt($curl, CURLOPT_HTTPHEADER, array('Authorization: Bearer ' . $token));

  $raindrop_results = json_decode(curl_exec($curl), true);
  curl_close($curl);
  return $raindrop_results;
}

// Function for rendering Raindrop.io collections in Alfred
function render_collections($raindrop_collections, $raindrop_collections_sublevel, $workflow, $render_style = "tree", $purpose = "adding", $parent_id = 0, $current_object = [], $current_level = -1)
{
  if ($parent_id == 0) {
    $collection_array = $raindrop_collections;
  } else {
    $collection_array = $raindrop_collections_sublevel;
  }
  foreach ($collection_array as $result) {
    if ($parent_id == 0 || $result["parent"]["\$id"] === $parent_id) {
      $current_level++;
      $current_object[$current_level] = $result["title"];

      $indentation = "";
      $sub_indentation = "";
      if ($render_style == "tree") {
        if ($current_level > 0) {
          $sub_indentation .= "\t";
        }
        for ($i = 1; $i < $current_level; $i++) {
          $indentation .= "\t";
          $sub_indentation .= "\t";
        }
        if ($current_level > 0) {
          $indentation .= "   ↳ ";
          $sub_indentation .= "   ";
        }
      }

      $icon_url_array = explode("/", $result["cover"][0]);
      if ($icon_url_array[key(array_slice($icon_url_array, -1, 1, true))] == "") {
        $icon_file_name = "folder.png";
      } else {
        $icon_file_name = "icon_cache/" . $icon_url_array[key(array_slice($icon_url_array, -1, 1, true))];
      }
      # Redownload the collection icon if the cached version is older than 60 days
      if (substr($icon_file_name, 0, 11) === "icon_cache/" && (!file_exists($icon_file_name) || time() - filemtime($icon_file_name) > 5184000)) {
        if (file_exists($icon_file_name)) {
          unlink($icon_file_name);
        }
        $icon_content = file_get_contents($result["cover"][0]);
        file_put_contents($icon_file_name, $icon_content);
      }

      $collection_title = $result["title"];
      if ($render_style == "paths") {
        $collection_title = implode("/", $current_object);
      }

      if ($purpose == "adding") {
        $workflow->result()
          ->arg($result["_id"] . " " . mb_strtolower(implode(" ", $current_object)) . " " . ($render_style == "tree" ? mb_strtolower(sub_collection_names($raindrop_collections_sublevel, $result["_id"])) : ""))
          ->mod('cmd', $sub_indentation . "Save now, without setting custom title or adding tags", "-↪︎" . $result["_id"] . " " . mb_strtolower(sub_collection_names($raindrop_collections_sublevel, $result["_id"])))
          ->icon($icon_file_name)
          ->title($indentation . $collection_title)
          ->mod("alt", "", $result["_id"] . " " . mb_strtolower(implode(" ", $current_object)) . " " . ($render_style == "tree" ? mb_strtolower(sub_collection_names($raindrop_collections_sublevel, $result["_id"])) : ""));
      } else if ($purpose == "searching") {
        $workflow->result()
          ->arg("⏦" . implode("/", $current_object) . "⏦" . $result["_id"] . "⏦" . $icon_file_name . "⏦" . mb_strtolower(implode(" ", $current_object)) . " " . ($render_style == "tree" ? mb_strtolower(sub_collection_names($raindrop_collections_sublevel, $result["_id"])) : ""))
          ->icon($icon_file_name)
          ->title($indentation . $collection_title)
          ->mod("alt", "", "⏦" . implode("/", $current_object) . "⏦" . $result["_id"] . "⏦" . $icon_file_name . "⏦" . mb_strtolower(implode(" ", $current_object)) . " " . ($render_style == "tree" ? mb_strtolower(sub_collection_names($raindrop_collections_sublevel, $result["_id"])) : ""));
      }

      render_collections($raindrop_collections, $raindrop_collections_sublevel, $workflow, $render_style, $purpose, $result["_id"], $current_object, $current_level);

      unset($current_object[$current_level]);
      $current_level--;
    }
  }
}

// Function for getting the names of all sub collections in a string
function sub_collection_names($raindrop_collections_sublevel, $parent_id)
{
  $names = "";
  foreach ($raindrop_collections_sublevel as $result) {
    if ($result["parent"]["\$id"] === $parent_id) {
      $names .= $result["title"] . " " . sub_collection_names($raindrop_collections_sublevel, $result["_id"]);
    }
  }
  return $names;
}

// Function for getting an array of all full path names for the collections, with the collection id's as keys.
function collection_paths($raindrop_collections, $raindrop_collections_sublevel, $path_list = [], $parent_id = 0, $current_object = [], $current_level = -1) {
  if ($parent_id == 0) {
    $collection_array = $raindrop_collections;
  } else {
    $collection_array = $raindrop_collections_sublevel;
  }
  foreach ($collection_array as $result) {
    if ($parent_id == 0 || $result["parent"]["\$id"] === $parent_id) {
      $current_level++;
      $current_object[$current_level] = $result["title"];
      $path_list[$result["_id"]] = implode("/", $current_object);

      $path_list = collection_paths($raindrop_collections, $raindrop_collections_sublevel, $path_list, $result["_id"], $current_object, $current_level);

      unset($current_object[$current_level]);
      $current_level--;
    }
  }
  return $path_list;
}

// Function for getting Raindrop.io collections
function collections(string $token, bool $sublevel, $caching = "check")
{
  // If $caching == "check": Redownload collection list only if cache is older than 1 minute, to make searching faster while still not having to wait for new collections to appear
  // If $caching == "trust": Trust the collection list cache to be good enough and use what is cached without checking its age (only download if no chache exists yet)
  // If $caching == "fetch": Always redownload collection list without checking the age of the cache
  if (file_exists($sublevel ? "collections_sublevel.json" : "collections.json") && ((time() - filemtime($sublevel ? "collections_sublevel.json" : "collections.json") < 60 && $caching == "check") || $caching == "trust")) {
    // Read stored cached collections
    $raindrop_results = json_decode(file_get_contents($sublevel ? "collections_sublevel.json" : "collections.json"), true);
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
  file_put_contents($sublevel ? "collections_sublevel.json" : "collections.json", json_encode($raindrop_results));
  return $raindrop_results;
}

// Function for getting Raindrop.io tags
function tags(string $token, $caching = "check")
{
  // If $caching == "check": Redownload tag list only if cache is older than 1 minute, to make searching faster while still not having to wait for new tags to appear
  // If $caching == "trust": Trust the tag list cache to be good enough and use what is cached without checking its age (only download if no chache exists yet)
  // If $caching == "fetch": Always redownload tag list without checking the age of the cache
  if (file_exists("tags.json") && ((time() - filemtime("tags.json") < 60 && $caching == "check") || $caching == "trust")) {
    // Read stored cached tags
    $raindrop_results = json_decode(file_get_contents("tags.json"), true);
    if ($raindrop_results["result"] == 1) {
      return $raindrop_results;
    }
  }

  // Query Raindrop.io
  $curl = curl_init();
  curl_setopt($curl, CURLOPT_URL, "https://api.raindrop.io/rest/v1/tags/0");
  curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
  curl_setopt($curl, CURLOPT_USERAGENT, "Alfred (Macintosh; Mac OS X)");
  curl_setopt($curl, CURLOPT_HTTPHEADER, array('Authorization: Bearer ' . $token));
  $raindrop_results = json_decode(curl_exec($curl), true);
  curl_close($curl);
  file_put_contents("tags.json", json_encode($raindrop_results));
  return $raindrop_results;
}