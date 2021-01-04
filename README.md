# Alfred Raindrop.io Search
This [Alfred](https://www.alfredapp.com) workflow is created to have a single, fast, and always reachable way of searching [Raindrop.io](https://raindrop.io), and then open bookmarks in the web browser you are currently working in, even if that's not your default browser.
You can also use this workflow to add bookmarks to [Raindrop.io](https://raindrop.io) from the browser you are currently using.

- To search your Raindrop.io bookmarks, open Alfred, type **r**, space, and then your search query, and the results will show directly in Alfred so that you can select one and press enter to open it in your browser.
  - Raindrop.io collections and tags will also show in the search results together with bookmarks, and you can select them to browse or search their content.
  - Before you have started to type a search query, you also have the option to browse your collections instead of starting with a search.
  - If a web browser is the frontmost app when you open a bookmark from this workflow, it will open in that browser.
  - If you are working in another app, the bookmark will open in your default browser.
  - Hold the cmd-key to view the URL for a bookmark.
  - Hold the ctrl-key to view the description for a bookmark.
  - Hold the option-key while pressing enter, or use cmd+c to copy the URL instead of opening it in a browser.
  - Press enter before you have started typing a search query, and Raindrop.io itself will open in your active web browser.
- To add a new bookmark to Raindrop.io, first make sure that you have the webpage you want to add opened in a browser and that it is the frontmost window, and then open Alfred and type **ra** followed by a space
  - In the first step you choose collection for the new bookmark, and you can either type to search for the collection you want to add the new bookmark to or just select one in the list. Hold the cmd-key to save when you select the collection, and skip setting a custom title or adding tags.
  - In the second step you get to change the title that the bookmark is saved with. Hold the cmd-key to save and skip the tag adding step.
  - In the third step you get to add tags to your new bookmark. You can either simply type them out, or select from a list of tags that matches what you have started to type. Separate multiple tags with comma. Hold the cmd-key to save when selecting a tag in the list, and skip the option of adding more tags.
  - The Firefox support for adding bookmarks was made possible with the help of deanishe's great workflow Firefox Assistant, which needs to be installed in Alfred for the Firefox support to function. Get it here: https://github.com/deanishe/alfred-firefox
- If the workflow is not authenticated with Raindrop.io when you initiate it, you will be taken to the authentication process.


## Supported Web Browsers
Safari, Chrome, Firefox, Edge, Brave, Vivaldi, Opera, Chromium, Chrome Canary, Safari Technology Preview, NAVER Whale (and SeaMonkey, but only for opening bookmarks)

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
Get your own codes at the following link and put them in client_code.json, and everything will work perfectly:
https://app.raindrop.io/#/settings/apps/dev