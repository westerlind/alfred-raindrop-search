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
$original_title = file_get_contents("current_title.tmp");

$workflow->result()
  ->arg($title)
  ->title("Save as \"" . $title . "\"")
  ->subtitle("Original title: " . $original_title);

// Output to Alfred
echo $workflow->output();