<?php
// Script for searching Raindrop.io bookmarks with Alfred,
// and handling parts of the OAuth authentication process
//
// By Andreas Westerlind in 2020
//

require './WorkflowsPHPHelper/src/Workflow.php';
require './WorkflowsPHPHelper/src/Result.php';
use Alfred\Workflows\Workflow;

$query = $argv[1];
$workflow = new Workflow;

// Check if the token file exists and otherwise send the user over to the authentication
if (!file_exists("token.json")) {
  init_auth ($workflow);
}

// Read token and related data from file
$token = json_decode(file_get_contents("token.json"), true);

if ($query != "") {
  // Query Raindrop.io
  $raindrop_results = search($query, $token["access_token"]);

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
      $raindrop_results = search($query, $new_token["access_token"]);
    }
  }

  // Prepare results for being viewed in Alfred
  foreach ($raindrop_results["items"] as $result) {
    $workflow->result()
      ->arg($result["link"])
      ->title($result["title"])
      ->subtitle($result["excerpt"] != "" ? $result["excerpt"] : $result["link"])
      ->copy($result["link"])
      ->mod('cmd', $result["link"], $result["link"]);
  }
}

if ($query == "") {
  // Check if Token has gone through more than half of it's lifetime, and in that case, refresh it
  $current_time = new DateTime("now", new DateTimeZone('UTC'));
  $token_time = date_create_from_format("Y-m-d H:i:s", $token["creation_time"], new DateTimeZone('UTC'));
  $time_difference = $token_time->diff($current_time);
  if ((int)$token["expires"] - date_interval_to_milliseconds($time_difference) < (int)$token["expires"] * 0.5) {
    refresh_token($token["refresh_token"]);
  }

  // Default results if nothing is searched for. Just go to Raindrop.io itself
  $workflow->result()
    ->arg("https://app.raindrop.io/")
    ->title("Search your Raindrop.io bookmarks")
    ->subtitle("Or press enter to open Raindrop.io");
}

// Output to Alfred
echo $workflow->output();

// ----------FUNCTIONS----------

// Function for initiating authentication process
function init_auth (Workflow $workflow) {
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
function refresh_token(string $refresh_token) {
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

// Function for searching Raindrop.io bookmarks
function search(string $query, string $token) {
  // Query Raindrop.io
  $curl = curl_init();
  curl_setopt($curl, CURLOPT_URL, "https://api.raindrop.io/rest/v1/raindrops/0/?search=[{\"key\":\"word\",\"val\":\"" . urlencode($query) . "\"}]");
  curl_setopt($curl, CURLOPT_RETURNTRANSFER, 1);
  curl_setopt($curl, CURLOPT_USERAGENT, "Alfred (Macintosh; Mac OS X)");
  curl_setopt($curl, CURLOPT_HTTPHEADER, array('Authorization: Bearer ' . $token));
  $raindrop_results = json_decode(curl_exec($curl), true);
  curl_close($curl);
  return $raindrop_results;
}

// Function for getting total amount of milliseconds from DateInterval variable
// Needed to not have to use any Unix timestamps, which should be avoided as much as possible.
function date_interval_to_milliseconds(DateInterval $interval) {
  $milliseconds = (int)$interval->format("%f") / 1000;
  $milliseconds += (int)$interval->format("%s") * 1000;
  $milliseconds += (int)$interval->format("%i") * 60 * 1000;
  $milliseconds += (int)$interval->format("%h") * 60 * 60 * 1000;
  $milliseconds += (int)$interval->format("%a") * 24 * 60 * 60 * 1000;
  return $milliseconds;
}