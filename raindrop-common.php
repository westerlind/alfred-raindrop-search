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
