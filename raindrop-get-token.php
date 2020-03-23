<?php
// Script for trading an OAuth authentication code for an authentication token from Raindrop.io
//
// By Andreas Westerlind in 2020
//

require 'raindrop-common.php';

$code = $argv[1];

// Get client_id and client_secret from client_code.json
// Those values are specific to this Alfred Workflow.
// Please get your own codes from https://app.raindrop.io/#/settings/apps/dev if you are going to use this script for something else!
$client_code = json_decode(file_get_contents("client_code.json"), true);

// Prepare POST variables
$post_variables = array(
  "code" => $code,
  "client_id" => $client_code["client_id"],
  "client_secret" => $client_code["client_secret"],
  "redirect_uri" => "http://127.0.0.1:11038",
  "grant_type" => "authorization_code");

// Query Raindrop.io for token
$curl = curl_init();
curl_setopt($curl, CURLOPT_URL, "https://raindrop.io/oauth/access_token");
curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
curl_setopt($curl, CURLOPT_USERAGENT, "Alfred (Macintosh; Mac OS X)");
curl_setopt($curl, CURLOPT_POSTFIELDS, http_build_query($post_variables));
$token = json_decode(curl_exec($curl), true);
curl_close($curl);

if($token["error"]) {
  echo "failure";
}
else {
  $creation_time = new DateTime("now", new DateTimeZone('UTC'));
  $token["creation_time"] = $creation_time->format("Y-m-d H:i:s");
  file_put_contents("token.json" , json_encode($token));
  echo "success";
}