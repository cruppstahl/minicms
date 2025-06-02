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
[ ] review templating - do we have enough data for a header/footer?
[ ] remove Directory and File structure in Navigation

[ ] migrate crupp.de to the new solution; objective is that this becomes
    (with minor modifications) the default template for this use case!
    [ ] add a /now page
    [ ] update config, navigation.yaml
    [ ] add a default favicon
    [ ] add the pdf (for downloading the CV)
    [ ] really use templating to add links etc
    [ ] automate the deployment, e.g. in a docker container
    [ ] cmd line args ("create --template=business-card-01") then copy this
        template to a new (clean!) subdirectory!

[ ] use case: host technical documentation
    [ ] self-host the documentation

[ ] use case: personal blog
    [ ] parse navigation.yaml and use it to build routes
    [ ] default index page shows all posts (configurable!)
    [ ] site configuration specifies pagination etc

[ ] if a file was updated, i.e. has a newer timestamp (and different checksum):
    rebuild it (but not more than once per minute)
[ ] if a file or directory was added then support the route
[ ] if a file or directory was removed then drop the route
[ ] make sure to also rebuild the generated pages etc
