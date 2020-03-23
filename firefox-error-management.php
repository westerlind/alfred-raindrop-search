<?php
// Script for managing errors with not being able to use Firefox Assistant for different reasons
//
// By Andreas Westerlind in 2020
//

require './WorkflowsPHPHelper/src/Workflow.php';
require './WorkflowsPHPHelper/src/Result.php';
require 'raindrop-common.php';
use Alfred\Workflows\Workflow;

$error = $argv[1];
$workflow = new Workflow;

//echo "THIS: \"". $error."\"\n";

if ($error == "Workflow with Id 'net.deanishe.alfred.firefox-assistant' is disabled") {
  $workflow->result()
    ->icon("firefox.png")
    ->valid(false)
    ->title("Firefox Assistant is disabled")
    ->subtitle("Enable it in Alfred's preferences to be able to add bookmarks from Firefox");
}
else if ($error == "Cannot find workflow with Id 'net.deanishe.alfred.firefox-assistant'") {
  $workflow->result()
    ->icon("firefox.png")
    ->arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md")
    ->title("Firefox Assistant is needed to add from Firefox")
    ->subtitle("Press enter for instructions to install and configure Firefox Assistant");
}
else if ($error == "Cannot Connect to Extension") {
  $workflow->result()
    ->icon("firefox.png")
    ->arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md")
    ->title("Cannot Connect to Firefox Extension")
    ->subtitle("Press enter for instructions to install and configure Firefox Assistant");
}
else {
  $workflow->result()
    ->icon("firefox.png")
    ->arg("https://github.com/deanishe/alfred-firefox/blob/master/README.md")
    ->title("Something went wrong while connecting to Firefox")
    ->subtitle("Press enter for instructions to install and configure Firefox Assistant");
}

$workflow->variable('ff_error', "true");

// Output to Alfred
echo $workflow->output();

