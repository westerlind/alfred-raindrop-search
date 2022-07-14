# Alfred Raindrop.io Search
This [Alfred](https://www.alfredapp.com) workflow is created to have a single, fast, and always reachable way of searching [Raindrop.io](https://raindrop.io), and then open bookmarks in the web browser you are currently working in, even if that's not your default browser.
You can also use this workflow to add bookmarks to [Raindrop.io](https://raindrop.io) from the browser you are currently using, or from an address that you have copied from somewhere else.

- To search your Raindrop.io bookmarks, open Alfred, type **r**, space, and then your search query, and the results will show directly in Alfred so that you can select a bookmark and press enter to open it in your browser.
  - Raindrop.io collections and tags will also show in the search results together with bookmarks, and you can select them to browse or search their content.
  - Before you have started to type a search query, you also have the option to browse your collections instead of starting with a search.
  - If a web browser is the frontmost app when you open a bookmark from this workflow, it will open in that browser.
  - If you are working in another app, the bookmark will open in your default browser.
  - Hold the cmd-key to view the URL for a bookmark.
  - Hold the ctrl-key to view the description for a bookmark.
  - Hold the option-key and press enter, or use cmd+c to copy the URL instead of opening it in a browser.
  - Hold the shift-key and press enter to open the permanent copy that is stored at Raindrop.io (Requires a Raindrop.io Pro subscription to work)
  - Press enter before you have started typing a search query, and Raindrop.io itself will open in your active web browser.
- To add a new bookmark to Raindrop.io, there are two ways to get the actual bookmark you want to add into the workflow.
    - The primary way is to first make sure that you have the webpage you want to add opened in a browser and that it is the frontmost window, and then open Alfred and type **ra** followed by a space.
    - The alternative way, which only works if the frontmost application is not one of the supported browsers (as the primary method will be used then), is that you first copy an address that you want to add as a bookmark, and then open Alfred and type **ra** followed by a space.
  - In the first step you then choose a collection for the new bookmark, and you can either type to search for the collection you want to add the new bookmark to or just select one in the list. Hold the cmd-key to save when you select the collection, and skip setting a custom title or adding tags.
  - In the second step you get to change the title that the bookmark is saved with. Hold the cmd-key to save and skip the tag adding step.
  - In the third step you get to add tags to your new bookmark. You can either simply type them out, or select from a list of tags that matches what you have started to type. Separate multiple tags with comma. Hold the cmd-key to save when selecting a tag in the list, and skip the option of adding more tags.
  - The Firefox support for adding bookmarks was made possible with the help of deanishe's great workflow Firefox Assistant, which needs to be installed in Alfred for the Firefox support to function. The workflow will tell you about this when it is needed and direct you to instructions about what you need to do, but you can also get it in advance here: https://github.com/deanishe/alfred-firefox
- If the workflow is not authenticated with Raindrop.io when you initiate it, you will be taken to the authentication process.

## Supported macOS versions
This workflow is compatible with macOS 10.13 High Sierra and newer, and doesn't have any other external requirements (other than for Firefox bookmark adding support. See a few lines above here for that).
The reason for not supporting older versions than High Sierra is that Go 1.17 doesn't support older OS versions.
You can still run version 1.7 of this workflow on older macOS versions if you need to be able to do that, as all versions of this workflow prior to 2.0 is based on PHP and Python rather than Go.

## Supported Web Browsers
Safari, Chrome, Firefox, Edge, Brave, Vivaldi, Opera, Chromium, Chrome Canary, Safari Technology Preview, Arc, Sidekick NAVER Whale (and SeaMonkey and SigmaOS, but only for opening bookmarks)

## Settings
- To set keyboard shortcuts, go to the "Search Raindrop.io" workflow in the Alfred preferences and look in the top left corner, where you can set keyboard shortcuts for searching Raindrop.io, or for adding bookmarks.
- To change other settings, go to the "Search Raindrop.io" workflow in the Alfred preferences, and click the [𝒙] button in the top right corner, where you get descriptions of the options in the information view to the left, and set the options by changing the value of the variables to the right.

## Availability
The main place to download the latest release of this workflow is at Packal, here:
https://www.packal.org/workflow/search-raindropio

You can also get it from GitHub here:
https://github.com/westerlind/alfred-raindrop-search/releases

## If You Download the Source Code
If you download the source code from GitHub rather than the packaged .alfredworkflow file, the codes that are used to identify this as a client to Raindrop.io is not included, as they are supposed to be application specific.
Get your own codes at the following link and put them in client_code.go:
https://app.raindrop.io/#/settings/apps/dev
Here is an instruction for how to get everything ready and compiling the code:
- Install the XCode command line tools with this terminal command: `xcode-select --install`
- Install Homebrew. Instructions [here](https://brew.sh/).
- Install Go by running the following in the terminal after installing Homebrew: `brew install go`
- You should now be able to compile the code by simply running the provided `build.sh`
- If that doesn't work, start by checking that Go functions properly. If you get Go to function, `build.sh` should also work.
- The result will be a universal binary (native for both Intel and Apple Silicon), with the name `raindrop_alfred`
