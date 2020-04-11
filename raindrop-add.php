<?php
// Script for adding a new bookmark to Raindrop.io
//
// By Andreas Westerlind in 2020
//

require 'raindrop-common.php';

// Get information about new bookmark
$selection = json_decode(file_get_contents("current_selection.tmp"), true);
$collection = $selection["collection"];
$url = $selection["url"];
$title = $selection["title"];
$tags = explode(",", $argv[1]);
foreach ($tags as $key => $current_tag) {
  $tags[$key] = trim(trim($current_tag), "#");
}

// Read token and related data from file.
// We assume that this exists, as it would not be possible to get here from within Alfred otherwise.
$token = json_decode(file_get_contents("token.json"), true);

// Add bookmark to Raindrop.io
raindrop_add($token["access_token"], $collection, $url, $title, $tags);

echo $title;

// ----------FUNCTIONS----------

// Function for adding a new bookmark to Alfred
function raindrop_add(string $token, string $collection, string $url, string $title, array $tags) {

  // Get meta description from the webpage we are adding, and use that as description for the bookmark.
  // Alfred is not really all that good for editing this sort of longer text, so the user will have to go to 
  // Raindrop.io and edit it there to have a custom description, or to get a description if there is no meta description tag.
  // This is much better than no ability to have a description at all though!
  $meta_tags = get_meta_tags($url);
  $description = "";
  if (isset($meta_tags["description"])) {
    $description = $meta_tags["description"];
  }

  // Prepare POST variables
  $post_variables = array(
    "collection" => array(
      "\$ref" => "collections",
      "\$id" => $collection
    ),
    "link" => $url,
    "title" => $title,
    "tags" => $tags,
    "excerpt" => $description
  ); 

  // Add to Raindrop.io
  $curl = curl_init();
  curl_setopt($curl, CURLOPT_URL, "https://api.raindrop.io/rest/v1/raindrop");
  curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
  curl_setopt($curl, CURLOPT_USERAGENT, "Alfred (Macintosh; Mac OS X)");
  curl_setopt($curl, CURLOPT_HTTPHEADER, array('Authorization: Bearer ' . $token, "Content-Type: application/json; charset=UTF-8", "X-Accept: application/json"));
  curl_setopt($curl, CURLOPT_POSTFIELDS, json_encode($post_variables));
  $raindrop_results = json_decode(curl_exec($curl), true);
  curl_close($curl);
  return $raindrop_results;
}