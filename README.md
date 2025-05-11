# Alfred Raindrop.io Search & Bookmark Management Workflow
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
- If you prefer faster searches over full-text search and more accurate search results, there is an alternative search mechanism for this. Open Alfred, type **rl**, space, and then your search query. This provides considerably faster results by searching a local cache of your bookmarks instead of querying the Raindrop.io API each time.
  - This search mechanism will search the title, excerpt/description, and link address of each bookmark, but full-text search is not supported with this mechanism, and depending on how you use Raindrop.io, the quality of the results may not be entirely as good.
  - The local cache is updated automatically the first time you do a local search after the configured update interval has passed (default 24h). The cache is refreshed after providing the bookmarks to Alfred for doing the current search, which means that you get your results as fast as possible, and the local cache is updated for the next search you search.
  - To manually refresh the local cache, open Alfred and type **rr**.
  - Other than full-text search, all the same features are available in the local search.
- As both of the search modes are available in parallel, you can, for example, assign them to different keyboard shortcuts and use the one that is better for the current purpose (either with full-text search or faster)
- If you prefer the faster local search over the full-text search capability, you can change the local search to use **r** in the workflow view (look for the green objects there) to keep it as easily available as possible. 
- To add a new bookmark to Raindrop.io, there are two ways to get the actual bookmark you want to add into the workflow.
    - The primary way is to first make sure that you have the webpage you want to add opened in a browser and that it is the frontmost window, and then open Alfred and type **ra** followed by a space.
    - The alternative way, which only works if the frontmost application is not one of the supported browsers (as the primary method will be used then), is that you first copy an address that you want to add as a bookmark, and then open Alfred and type **ra** followed by a space.
  - In the first step you then choose a collection for the new bookmark, and you can either type to search for the collection you want to add the new bookmark to or just select one in the list. Hold the cmd-key to save when you select the collection, and skip setting a custom title or adding tags.
  - In the second step you get to change the title that the bookmark is saved with. Hold the cmd-key to save and skip the tag adding step.
  - In the third step you get to add tags to your new bookmark. You can either simply type them out, or select from a list of tags that matches what you have started to type. Separate multiple tags with comma. Hold the cmd-key to save when selecting a tag in the list, and skip the option of adding more tags.
  - The Firefox support for adding bookmarks was made possible with the help of deanishe's great workflow Firefox Assistant, which needs to be installed in Alfred for the Firefox support to function. The workflow will tell you about this when it is needed and direct you to instructions about what you need to do, but you can also get it in advance here: https://github.com/deanishe/alfred-firefox
- If the workflow is not authenticated with Raindrop.io when you initiate it, you will be taken to the authentication process.
- You can log out from Raindrop.io by opening Alfred and typing rlogout

## Supported macOS versions
This workflow is compatible with macOS 11 Big Sur and newer, and doesn't have any other external requirements (other than for Firefox bookmark adding support. See a few lines above here for that).
The reason for not supporting older versions than Big Sur is that Go 1.24 doesn't support older OS versions.

## Supported Alfred versions
Alfred 5 is supported from version 2.0.7 onwards and is the only supported version in 3.X.
Alfred 4 is supported for all versions before 3.0.

## Supported Web Browsers
Safari, Chrome, Firefox, Edge, Brave, Vivaldi, Opera, Arc, Orion, Sidekick, Chromium, Chrome Canary, Safari Technology Preview, NAVER Whale (and SeaMonkey and SigmaOS, but only for opening bookmarks)

## Settings
- To set keyboard shortcuts, go to the "Search Raindrop.io" workflow in the Alfred preferences and look in the top left corner where you can set keyboard shortcuts for searching Raindrop.io, searching the local cache, or for adding bookmarks.
- To change other settings, go to the "Search Raindrop.io" workflow in the Alfred preferences, and click the "Configure Workflow" button above the workflow view, where you can change the settings and se descriptions of what each setting is doing.

## Availability
You can download the newest version of this workflow from GitHub here:
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
