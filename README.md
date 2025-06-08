[x] create scaffold for a golang project with gin
[x] if argv[1] == "run" then argv[2] is the directory with the data
[x] if argv[1] == "help" then print the help screen
[x] read authors.yaml - must not be empty
[x] read global configuration file
[x] use cmd line arguments to overwrite config values
[x] move everything to a github repository
[x] read files and .yaml metadata in each directory, descend recursively
    (do not read the files though)
[x] dynamically build all routes to the directories in "/content"
[x] Bug: do not create routes for subdirectories (e.g. /blog/2025)
[x] Bug: all URLs return the same content (post3.html)
[x] build page for a file on demand and in case it was not yet created
[x] store the generated file in a cache
[x] fetch the generated file from a cache

[x] assemble the layout (header and footer)
[x] return html instead of json, with the correct mimetype
[x] use picocss for a basic layout
[x] use go template functions to complete header and footer

[x] parse navigation.yaml into the Context structure
    [x] for each top level item in the navigation: create a route
    [x] repeat for each nested item (in the navigation)
    [x] for each *file* in the navigation directory: create a route
    [x] for each *subdir* in the navigation directory: create a route
    [x] and store the meta-information in the navigation directory
    [x] "/" in a directory redirects to the index page
    [x] enforce absolute paths (urls and locals) in navigation.yaml
[x] DataTree is then no longer required
[x] Hide file extension in the route (e.g. /index.md -> /index)

[x] add favicon to site configuration (under "branding")
[x] add route for static files (router.Static("/static", "./local-assets")
[x] review templating - do we have enough data for a header/footer?
	copy Navigation, (parts of)Config, Branding, Users
	also copy relevant stuff from Directory, File (merge them - settings in
	File have higher priority than those in Directory)
[x] move custom css template (site.css) to Branding, use it in header.html
	adjust header.html, footer.html
[x] remove Directory and File structure in Navigation; Directory is only
	required from the File struct itself (as a reference), but it is
	not required as a standalone object
[x] File structure is only required in the LookupIndex
[x] make File.CachedContent a byte array, not a string
[x] if the data tree changes, then all dependent files need to be regenerated
	[x] $site/layout: invalidate everything (i.e. delete CachedContent of all
        text/html files)
	[x] directory: everything including/below that directory is
        invalidated
	[x] directory metadata: everything including/below that directory is
        invalidated
	[x] file metadata: file is invalidated
	[x] file: file is recreated

[ ] Create a default template for a minimalistic digital business card
    [x] Create a new layout for the new page
    [x] break it up into multiple html files
    [x] Move inline css to separate file
    [x] update config, navigation.yaml
    [x] update the main page (CV) if necessary
    [x] display date of last update (of the current page) in the footer
    [x] add a favicon (default symbol: • or ·)
    [x] use templating to add title, description
    [x] Check css and html with a linter, and format them properly
    [x] use templating to add navigation links
    [ ] cmd line args ("create --template=business-card-01") then copy this
        template to a new (clean!) subdirectory!
    [ ] migrate crupp.de to the new solution
    	[ ] automate the deployment, e.g. in a docker container
    	[ ] set up monitoring

[ ] use case: host technical documentation
    [ ] self-host the documentation of what we have built so far
    [ ] Support markdown templating and formatting
    [ ] Source code is formatted in a different style, with syntax highlighting
    [ ] Add search functionality, with key words and full text

[ ] use case: personal blog
    [ ] parse navigation.yaml and use it to build routes
    [ ] default index page shows all posts (configurable!)
    [ ] site configuration specifies pagination etc
    [ ] recreate auto-generated blog file if a post has been changed or
        added/removed

[ ] if a file or directory was added then support the route
[ ] if a file or directory was removed then drop the route
