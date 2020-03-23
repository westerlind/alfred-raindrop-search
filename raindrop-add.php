<?php
// Script for adding a new bookmark to Raindrop.io
//
// By Andreas Westerlind in 2020
//

require 'raindrop-common.php';

// Get variables from Alfred.
// The §+#^§ divider thing is a workaround for issues with passing a parameter to a script within Alfred.
// It doesn't look very good, but it works.
$category = explode(" ", $argv[1], 2)[0];
$url = explode("§+#^§", $argv[1])[1];
$title = explode("§+#^§", $argv[1])[2];

// Read token and related data from file.
// We assume that this exists, as it would not be possible to get here from within Alfred otherwise.
$token = json_decode(file_get_contents("token.json"), true);

// Add bookmark to Raindrop.io
raindrop_add($token["access_token"], $category, $url, $title);

// ----------FUNCTIONS----------

// Function for adding a new bookmark to Alfred
function raindrop_add(string $token, string $category, string $url, string $title) {
  
  // Prepare POST variables
  $post_variables = array(
    "collection" => array(
      "\$ref" => "collections",
      "\$id" => $category
    ),
    "link" => $url,
    "title" => $title
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