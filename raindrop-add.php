<?php
// Script for adding a new bookmark to Raindrop.io
//
// By Andreas Westerlind in 2020
//

require 'raindrop-common.php';

// Get information about new bookmark
$selection  = file_get_contents("current_selection.tmp");
$collection = explode(" ", $selection)[0];
$url = explode("§+#^§", $selection)[1];
$title = file_get_contents("current_title.tmp");
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

  // Prepare POST variables
  $post_variables = array(
    "collection" => array(
      "\$ref" => "collections",
      "\$id" => $collection
    ),
    "link" => $url,
    "title" => $title,
    "tags" => $tags
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