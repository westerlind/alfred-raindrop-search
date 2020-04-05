# Alfred Raindrop.io Search
This Alfred Workflow is created to have a single, fast, and always reachable way of searching Raindrop.io, and then open bookmarks in the web browser you are currently working in, even if that's not your default browser.
You can also add bookmarks to Raindrop.io from the browser you are currently using.

- To search your Raindrop.io bookmarks, open Alfred, type **r**, space, and then your search query, and the results will show directly in Alfred so that you can select one and press enter to open it in your browser.
  - If a web browser is the frontmost app when you open a bookmark from this workflow, it will open in that browser.
  - If you are working in another app, the bookmark will open in your default browser.
  - Hold the cmd-key to view the URL for a bookmark.
  - Hold the option-key while pressing enter, or use cmd+c to copy the URL instead of opening it in a browser.
  - Press enter before you have started typing a search query, and Raindrop.io itself will open in your active web browser.
- To add a new bookmark to Raindrop.io, first make sure that you have the page you want to add opened in a browser and that it is the frontmost window and then open Alfred, type **ra**, space, and either type to search for the collection you want to add the new bookmark to or just select one in the list.
  - The Firefox support for adding bookmarks was made possible with the help of deanishe's great workflow Firefox Assistant, which needs to be installed in Alfred for the Firefox support to function. Get it here: https://github.com/deanishe/alfred-firefox
  - Hold the cmd-key while selecting a collection to open Raindrop.io in your active web browser after saving, so you can change details of your new bookmark there.
- If the workflow is not authenticated with Raindrop.io when you initiate it, you will be taken to the authentication process.
- Go to the "Search Raindrop.io" workflow in the Alfred preferences and look in the top left corner if you want to add a keyboard shortcuts for going directly to the Raindrop.io search or bookmark adding.

## Supported Web Browsers
Safari, Chrome, Firefox, Edge, Brave, Vivaldi, Opera, Chromium, Chrome Canary, Safari Technology Preview (and SeaMonkey, but only for opening bookmarks)

## Availability
The main place to download the latest release of this workflow is at Packal, here:
https://www.packal.org/workflow/search-raindropio

You can also get it from GitHub here:
https://github.com/westerlind/alfred-raindrop-search/releases

## If You Download the Source Code
If you download the source code from GitHub rather than the packaged .alfredworkflow file, the codes that are used to identify this as a client to Raindrop.io is not included, as they are supposed to be application specific.
Get your own codes at the following link and put them in client_code.json, and everything will work perfectly:
https://app.raindrop.io/#/settings/apps/dev