<?php
// Script for supporting the process of setting the title for a
// new bookmark that is added to Raindrop.io via Alfred
//
// By Andreas Westerlind in 2020
//

require './WorkflowsPHPHelper/src/Workflow.php';
require './WorkflowsPHPHelper/src/Result.php';
require 'raindrop-common.php';

use Alfred\Workflows\Workflow;

$workflow = new Workflow;

$title = $argv[1];
$original_title = json_decode(file_get_contents("current_selection.tmp"), true)["title"];

$workflow->result()
  ->arg($title)
  ->title("Save as: " . $title)
  ->subtitle("Original title: " . $original_title)
  ->mod('cmd', "Save now, without adding tags", "-↪︎" . $title)
  ->mod("alt", "Original title: " . $original_title, $title);

// Output to Alfred
echo $workflow->output();